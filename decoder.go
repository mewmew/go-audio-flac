package flac

import (
	"io"

	"github.com/go-audio/audio"
	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/frame"
)

// Decoder handles the decoding of FLAC files.
type Decoder struct {
	stream *flac.Stream
	frame  *frame.Frame // current frame.
	i      int          // index of current sample in subframe(s).

	// TODO: provide access to metadata.
	// Metadata for the current file.
	//Metadata *Metadata
}

// NewDecoder creates a decoder for the passed FLAC reader.
func NewDecoder(r io.Reader) (*Decoder, error) {
	stream, err := flac.New(r)
	if err != nil {
		return nil, err
	}
	d := &Decoder{
		stream: stream,
	}
	return d, nil
}

// TODO: implement support for Seek.
// Seek provides access to the cursor position in the PCM data.
//func (d *Decoder) Seek(offset int64, whence int) (int64, error) {
//	return d.r.Seek(offset, whence)
//	panic("flac.Decoder.Seek: not yet implemented")
//}

// SampleBitDepth returns the bit depth encoding of each sample.
func (d *Decoder) SampleBitDepth() int32 {
	if d == nil {
		return 0
	}
	return int32(d.stream.Info.BitsPerSample)
}

// PCMBuffer populates the passed PCM buffer.
func (d *Decoder) PCMBuffer(buf *audio.IntBuffer) (n int, err error) {
	if buf == nil {
		return 0, nil
	}
	buf.SourceBitDepth = int(d.SampleBitDepth())
	buf.Format = d.Format()

	// Fill b with audio samples from the previous decoded audio frame.
	if d.frame != nil {
		for ; d.i < int(d.frame.BlockSize); d.i++ {
			for _, subframe := range d.frame.Subframes {
				sample := subframe.Samples[d.i]
				if n >= len(buf.Data) {
					return n, nil
				}
				buf.Data[n] = int(sample)
				n++
			}
		}
	}
	d.frame = nil

	// Fill b with audio samples from decoded audio frames.
	for {
		frame, err := d.stream.ParseNext()
		if err != nil {
			if err == io.EOF {
				return n, io.EOF // TODO: signal end of stream differently?
			}
			return n, err
		}
		for i := 0; i < int(frame.BlockSize); i++ {
			for _, subframe := range frame.Subframes {
				sample := subframe.Samples[i]
				if n >= len(buf.Data) {
					if i != int(frame.BlockSize)-1 {
						// Fewer audio samples were read than contained within the audio
						// frame. Store the decoded audio frame and the current sample
						// position for future read operations.
						d.frame = frame
						d.i = i
					}
					return n, nil
				}
				buf.Data[n] = int(sample)
				n++
			}
		}
	}
}

// Format returns the audio format of the decoded content.
func (d *Decoder) Format() *audio.Format {
	if d == nil {
		return nil
	}
	return &audio.Format{
		NumChannels: int(d.stream.Info.NChannels),
		SampleRate:  int(d.stream.Info.SampleRate),
	}
}

// Duration returns the time duration for the current audio container.
//func (d *Decoder) Duration() (time.Duration, error) {
//	if d == nil || d.parser == nil {
//		return 0, errors.New("can't calculate the duration of a nil pointer")
//	}
//	return d.parser.Duration()
//}

// Close closes the underlying FLAC stream of the decoder.
func (d *Decoder) Close() error {
	return d.stream.Close()
}
