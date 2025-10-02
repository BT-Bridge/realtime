package main

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"os"
	"os/signal"
	"syscall"
	"time"

	"encoding/binary"

	"github.com/gordonklaus/portaudio"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/realtime"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/valyala/fasthttp"
	"gitlab.bcc-hyperdev.org/bcc-hyperdev/realtime/shared"
	"go.uber.org/zap"
)

type OpenaiRealtimeClient struct {
	logger    *shared.Logger
	apiKey    string
	orgId     string
	projectId string
	baseUrl   string
}

func NewOpenaiRealtimeClient(logger *shared.Logger, apiKey, orgId, projectId, baseUrl string) (*OpenaiRealtimeClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey is required")
	}
	if orgId == "" {
		return nil, fmt.Errorf("orgId is required")
	}
	if projectId == "" {
		return nil, fmt.Errorf("projectId is required")
	}
	if baseUrl == "" {
		baseUrl = "https://api.openai.com/v1"
	}
	return &OpenaiRealtimeClient{
		logger:    logger,
		apiKey:    apiKey,
		orgId:     orgId,
		projectId: projectId,
		baseUrl:   baseUrl,
	}, nil
}

type OpenaiConfig struct {
	ApiKey    string
	OrgId     string
	ProjectId string
	BaseUrl   string
}

type OpenaiRealtimeService struct {
	logger *shared.Logger
	cfg    *OpenaiConfig
}

func NewOpenaiRealtimeService(logger *shared.Logger, cfg *OpenaiConfig) (s *OpenaiRealtimeService, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create OpenAI Realtime Service: %w", err)
		}
	}()
	return &OpenaiRealtimeService{
		logger: logger,
		cfg:    cfg,
	}, nil
}

func (s *OpenaiRealtimeService) NewClient() (c *OpenaiRealtimeClient, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create client: %w", err)
		}
	}()
	return NewOpenaiRealtimeClient(
		s.logger,
		s.cfg.ApiKey,
		s.cfg.OrgId,
		s.cfg.ProjectId,
		s.cfg.BaseUrl,
	)
}

func main() {
	logger := shared.NewLogger(
		zap.String("package", "realtime"),
		zap.String("example", "openai"),
	)
	svc, err := NewOpenaiRealtimeService(
		logger,
		&OpenaiConfig{
			ApiKey:    shared.MustGetenv(shared.GetenvString, "OPENAI_API_KEY", true),
			OrgId:     shared.MustGetenv(shared.GetenvString, "OPENAI_ORG_ID", true),
			ProjectId: shared.MustGetenv(shared.GetenvString, "OPENAI_PROJECT_ID", true),
			BaseUrl:   shared.MustGetenv(shared.GetenvString, "OPENAI_BASE_URL", false, "https://api.openai.com/v1"),
		},
	)
	if err != nil {
		logger.NoCtxFatal(err.Error())
	}
	client, err := svc.NewClient()
	_ = client
	if err != nil {
		logger.NoCtxFatal(err.Error())
	}
	request := realtime.RealtimeSessionCreateRequestParam{
		Instructions: param.NewOpt("You are a helpful assistant."),
		Model:        realtime.RealtimeSessionCreateRequestModelGPTRealtime,
		Audio: realtime.RealtimeAudioConfigParam{
			Input: realtime.RealtimeAudioConfigInputParam{
				TurnDetection: realtime.RealtimeAudioInputTurnDetectionUnionParam{
					OfSemanticVad: &realtime.RealtimeAudioInputTurnDetectionSemanticVadParam{
						CreateResponse:    param.NewOpt(true),
						InterruptResponse: param.NewOpt(true),
						Eagerness:         "low",
					},
				},
				Format: realtime.RealtimeAudioFormatsUnionParam{
					OfAudioPCM: &realtime.RealtimeAudioFormatsAudioPCMParam{
						Rate: 24000,
						Type: "audio/pcm",
					},
				},
				NoiseReduction: realtime.RealtimeAudioConfigInputNoiseReductionParam{
					Type: realtime.NoiseReductionTypeNearField,
				},
				Transcription: realtime.AudioTranscriptionParam{
					Language: param.NewOpt("fa"),
					Prompt:   param.NewOpt("expect words related to web technologies"),
					Model:    realtime.AudioTranscriptionModelWhisper1,
				},
			},
			Output: realtime.RealtimeAudioConfigOutputParam{
				Speed: param.NewOpt(0.9),
				Format: realtime.RealtimeAudioFormatsUnionParam{
					OfAudioPCM: &realtime.RealtimeAudioFormatsAudioPCMParam{
						Rate: 24000,
						Type: "audio/pcm",
					},
				},
				Voice: realtime.RealtimeAudioConfigOutputVoiceCedar,
			},
		},
		MaxOutputTokens: realtime.RealtimeSessionCreateRequestMaxOutputTokensUnionParam{
			OfInt: param.NewOpt[int64](1024),
		},
	}
	sessionConfig, err := request.MarshalJSON()
	if err != nil {
		logger.NoCtxFatal(fmt.Errorf("failed to marshal request: %w", err).Error())
	}
	fmt.Println("Session Config\n------------\n", string(sessionConfig))

	// Register Opus (48kHz stereo, but you can do mono too)
	me := &webrtc.MediaEngine{}
	err = me.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeOpus,
			ClockRate:   48000,
			Channels:    2, // or 1 if you want mono
			SDPFmtpLine: "minptime=10;useinbandfec=1",
		},
		PayloadType: 111, // Standard PT for Opus
	}, webrtc.RTPCodecTypeAudio)
	if err != nil {
		panic(err)
	}

	// Create a new API
	api := webrtc.NewAPI(webrtc.WithMediaEngine(me))

	// Create a new PeerConnection
	pc, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		panic(err)
	}

	// Create data channel with label OpenAI uses in examples
	dc, err := pc.CreateDataChannel("oai-events", nil)
	if err != nil {
		panic(err)
	}

	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		fmt.Printf("Received message on data channel: %s\n", string(msg.Data))
	})

	dc.OnOpen(func() {
		fmt.Println("Data channel opened.")
	})
	// Also accept incoming remote data channels (defensive)
	pc.OnDataChannel(func(remote *webrtc.DataChannel) {
		fmt.Printf("Remote data channel opened: %s\n", remote.Label())
		remote.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("Message from remote (%s): %s\n", remote.Label(), string(msg.Data))
		})
	})

	// Handle data channel messages if any
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		fmt.Printf("Received message on data channel: %s\n", string(msg.Data))
	})

	dc.OnOpen(func() {
		fmt.Println("Data channel opened.")
		// No need to send audio here; using RTP
		// Optionally send initial event if required
	})

	// Add transceiver for audio sendrecv
	_, err = pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendrecv,
	})
	if err != nil {
		panic(err)
	}

	// Initialize PortAudio
	err = portaudio.Initialize()
	if err != nil {
		panic(err)
	}
	defer func() { _ = portaudio.Terminate() }()

	const sampleRate = 24000
	const channels = 1
	const frameSize = 480 // 20ms at 24kHz
	inputBuffer := make([]int16, frameSize)
	outputBuffer := make([]int16, frameSize)

	// Open microphone stream
	inputStream, err := portaudio.OpenDefaultStream(channels, 0, sampleRate, frameSize, inputBuffer)
	if err != nil {
		panic(err)
	}
	defer func() { _ = inputStream.Close() }()

	err = inputStream.Start()
	if err != nil {
		panic(err)
	}

	// Open speaker stream
	outputStream, err := portaudio.OpenDefaultStream(0, channels, sampleRate, frameSize, outputBuffer)
	if err != nil {
		panic(err)
	}
	defer func() { _ = outputStream.Close() }()

	err = outputStream.Start()
	if err != nil {
		panic(err)
	}

	// Create local track for sending mic audio
	audioTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{
			MimeType:  webrtc.MimeTypeOpus,
			ClockRate: 48000,
			Channels:  1, // mono mic
		},
		"audio",
		"pion",
	)
	if err != nil {
		panic(err)
	}

	_, err = pc.AddTrack(audioTrack)
	if err != nil {
		panic(err)
	}

	go func() {
		buf := make([]int16, 960) // 20ms @ 48kHz mono
		for {
			err := inputStream.Read()
			if err != nil {
				fmt.Println("mic read error:", err)
				return
			}

			// Convert int16 -> []byte little-endian PCM
			data := make([]byte, len(buf)*2)
			for i, s := range buf {
				data[2*i] = byte(s)
				data[2*i+1] = byte(s >> 8)
			}

			// Write as a media.Sample (timestamping handled by Pion)
			err = audioTrack.WriteSample(media.Sample{
				Data:     data,
				Duration: 20 * time.Millisecond,
			})
			if err != nil {
				fmt.Println("track write error:", err)
				return
			}
		}
	}()
	// Set a handler for when a new remote track starts
	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Printf("Track has started, of type %d: %s\n", track.PayloadType(), track.Codec().MimeType)

		go func() {
			for {
				pkt, _, readErr := track.ReadRTP()
				if readErr != nil {
					fmt.Printf("Error reading RTP: %v\n", readErr)
					return
				}

				// Convert big-endian payload to int16 for PortAudio
				for i := 0; i < len(pkt.Payload)/2; i++ {
					outputBuffer[i] = int16(binary.BigEndian.Uint16(pkt.Payload[2*i:]))
				}

				writeErr := outputStream.Write()
				if writeErr != nil {
					fmt.Printf("Error writing to speaker: %v\n", writeErr)
					return
				}
			}
		}()
	})

	// Create offer
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	// Set the local description
	err = pc.SetLocalDescription(offer)
	if err != nil {
		panic(err)
	}

	// Wait for ICE gathering to complete so the offer includes candidates.
	// This mirrors typical browser behavior where offers include candidates
	// or use trickle ICE. Pion provides GatheringCompletePromise.
	<-webrtc.GatheringCompletePromise(pc)
	local := pc.LocalDescription()
	if local != nil {
		offer = *local
	}

	// Create multipart form data
	bodyBuffer := new(bytes.Buffer)
	writer := multipart.NewWriter(bodyBuffer)

	// For sdp with custom Content-Type
	sdpHeaders := textproto.MIMEHeader{}
	sdpHeaders.Set("Content-Disposition", `form-data; name="sdp"`)
	sdpHeaders.Set("Content-Type", "application/sdp")
	sdpPart, err := writer.CreatePart(sdpHeaders)
	if err != nil {
		panic(err)
	}
	_, err = sdpPart.Write([]byte(offer.SDP))
	if err != nil {
		panic(err)
	}

	// For session with custom Content-Type
	sessionHeaders := textproto.MIMEHeader{}
	sessionHeaders.Set("Content-Disposition", `form-data; name="session"`)
	sessionHeaders.Set("Content-Type", "application/json")
	sessionPart, err := writer.CreatePart(sessionHeaders)
	if err != nil {
		panic(err)
	}
	_, err = sessionPart.Write([]byte(sessionConfig))
	if err != nil {
		panic(err)
	}

	err = writer.Close()
	if err != nil {
		panic(err)
	}

	// Send request to OpenAI
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI("https://api.openai.com/v1/realtime/calls")
	req.Header.SetMethod("POST")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
	// req.Header.Set("OpenAI-Beta", "realtime=v1")
	req.Header.SetContentType(writer.FormDataContentType())
	req.SetBody(bodyBuffer.Bytes())

	fmt.Println("Full Http Request\n------------")
	fmt.Println(string(req.Header.Header()))
	fmt.Println(string(req.Body()))

	err = fasthttp.Do(req, resp)
	if err != nil {
		panic(fmt.Sprintf("Failed to send request to OpenAI: %v", err))
	}

	if resp.StatusCode() != fasthttp.StatusCreated {
		panic(fmt.Sprintf("OpenAI returned status %d: %s", resp.StatusCode(), string(resp.Body())))
	}

	answerSDP := string(resp.Body())
	fmt.Println("Received SDP answer from OpenAI:\n", answerSDP)

	select {}

	// Set remote description

	// Wait for interrupt to stop

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("Shutting down...")
}

// TODO: support other fields
// type RealtimeSessionCreateRequestParam struct {
// 	Prompt responses.ResponsePromptParam `json:"prompt,omitzero"`
// 	Tracing RealtimeTracingConfigUnionParam `json:"tracing,omitzero"`
// 	Include []string `json:"include,omitzero"`
// 	OutputModalities []string `json:"output_modalities,omitzero"`
// 	ToolChoice RealtimeToolChoiceConfigUnionParam `json:"tool_choice,omitzero"`
// 	Tools RealtimeToolsConfigParam `json:"tools,omitzero"`
// 	Truncation RealtimeTruncationUnionParam `json:"truncation,omitzero"`
// }
