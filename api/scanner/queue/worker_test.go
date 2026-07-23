package queue

import (
	"testing"

	"gopkg.in/vansante/go-ffprobe.v2"
)

func TestBuildVideoMetadata(t *testing.T) {
	videoStream := func(avgFrameRate string) *ffprobe.Stream {
		return &ffprobe.Stream{
			CodecType:     string(ffprobe.StreamVideo),
			CodecLongName: "H.264",
			Width:         1920,
			Height:        1080,
			AvgFrameRate:  avgFrameRate,
			BitRate:       "1000000",
			Profile:       "High",
		}
	}

	audioStream := func(channels int) *ffprobe.Stream {
		return &ffprobe.Stream{
			CodecType: string(ffprobe.StreamAudio),
			Channels:  channels,
		}
	}

	tests := []struct {
		name          string
		data          *ffprobe.ProbeData
		wantErr       bool
		wantAudio     string
		wantFramerate *float64
	}{
		{
			name: "no video stream at all",
			data: &ffprobe.ProbeData{
				Format:  &ffprobe.Format{},
				Streams: []*ffprobe.Stream{audioStream(2)},
			},
			wantErr: true,
		},
		{
			name: "video stream, no audio stream",
			data: &ffprobe.ProbeData{
				Format:  &ffprobe.Format{DurationSeconds: 12.5},
				Streams: []*ffprobe.Stream{videoStream("30/1")},
			},
			wantAudio:     "No audio",
			wantFramerate: ptr(30.0),
		},
		{
			name: "audio stream present but 0 channels",
			data: &ffprobe.ProbeData{
				Format:  &ffprobe.Format{},
				Streams: []*ffprobe.Stream{videoStream("25/1"), audioStream(0)},
			},
			wantAudio:     "No audio",
			wantFramerate: ptr(25.0),
		},
		{
			name: "mono audio",
			data: &ffprobe.ProbeData{
				Format:  &ffprobe.Format{},
				Streams: []*ffprobe.Stream{videoStream("25/1"), audioStream(1)},
			},
			wantAudio:     "Mono audio",
			wantFramerate: ptr(25.0),
		},
		{
			name: "stereo audio",
			data: &ffprobe.ProbeData{
				Format:  &ffprobe.Format{},
				Streams: []*ffprobe.Stream{videoStream("25/1"), audioStream(2)},
			},
			wantAudio:     "Stereo audio",
			wantFramerate: ptr(25.0),
		},
		{
			name: "surround audio, more than 2 channels",
			data: &ffprobe.ProbeData{
				Format:  &ffprobe.Format{},
				Streams: []*ffprobe.Stream{videoStream("25/1"), audioStream(6)},
			},
			wantAudio:     "Audio (6 channels)",
			wantFramerate: ptr(25.0),
		},
		{
			name: "fractional NTSC framerate",
			data: &ffprobe.ProbeData{
				Format:  &ffprobe.Format{},
				Streams: []*ffprobe.Stream{videoStream("30000/1001")},
			},
			wantAudio:     "No audio",
			wantFramerate: ptr(30000.0 / 1001.0),
		},
		{
			name: "empty AvgFrameRate string -> no framerate",
			data: &ffprobe.ProbeData{
				Format:  &ffprobe.Format{},
				Streams: []*ffprobe.Stream{videoStream("")},
			},
			wantAudio:     "No audio",
			wantFramerate: nil,
		},
		{
			name: "zero denominator -> no framerate",
			data: &ffprobe.ProbeData{
				Format:  &ffprobe.Format{},
				Streams: []*ffprobe.Stream{videoStream("30/0")},
			},
			wantAudio:     "No audio",
			wantFramerate: nil,
		},
		{
			name: "non-numeric AvgFrameRate -> no framerate",
			data: &ffprobe.ProbeData{
				Format:  &ffprobe.Format{},
				Streams: []*ffprobe.Stream{videoStream("abc/def")},
			},
			wantAudio:     "No audio",
			wantFramerate: nil,
		},
		{
			name: "malformed AvgFrameRate (wrong number of parts) -> no framerate",
			data: &ffprobe.ProbeData{
				Format:  &ffprobe.Format{},
				Streams: []*ffprobe.Stream{videoStream("30/1/1")},
			},
			wantAudio:     "No audio",
			wantFramerate: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildVideoMetadata(tc.data)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("buildVideoMetadata() error = nil, want an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("buildVideoMetadata() unexpected error: %v", err)
			}

			if got.Audio == nil || *got.Audio != tc.wantAudio {
				t.Errorf("Audio = %v, want %q", derefStr(got.Audio), tc.wantAudio)
			}

			switch {
			case tc.wantFramerate == nil && got.Framerate != nil:
				t.Errorf("Framerate = %v, want nil", *got.Framerate)
			case tc.wantFramerate != nil && got.Framerate == nil:
				t.Errorf("Framerate = nil, want %v", *tc.wantFramerate)
			case tc.wantFramerate != nil && got.Framerate != nil && *got.Framerate != *tc.wantFramerate:
				t.Errorf("Framerate = %v, want %v", *got.Framerate, *tc.wantFramerate)
			}

			wantStream := tc.data.FirstVideoStream()
			if got.Width != wantStream.Width || got.Height != wantStream.Height {
				t.Errorf("dimensions = %dx%d, want %dx%d", got.Width, got.Height, wantStream.Width, wantStream.Height)
			}
			if got.Duration != tc.data.Format.DurationSeconds {
				t.Errorf("Duration = %v, want %v", got.Duration, tc.data.Format.DurationSeconds)
			}
			if got.Codec == nil || *got.Codec != wantStream.CodecLongName {
				t.Errorf("Codec = %v, want %q", derefStr(got.Codec), wantStream.CodecLongName)
			}
			if got.Bitrate == nil || *got.Bitrate != wantStream.BitRate {
				t.Errorf("Bitrate = %v, want %q", derefStr(got.Bitrate), wantStream.BitRate)
			}
			if got.ColorProfile == nil || *got.ColorProfile != wantStream.Profile {
				t.Errorf("ColorProfile = %v, want %q", derefStr(got.ColorProfile), wantStream.Profile)
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }

func derefStr(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
