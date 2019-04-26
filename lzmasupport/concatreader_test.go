package lzmasupport

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConcatReader(t *testing.T) {
	assert := assert.New(t)

	makeCR := func() io.Reader {
		r1 := strings.NewReader("foo")
		r2 := strings.NewReader("bar")
		r3 := strings.NewReader("baz")
		return &concatReader{readers: []io.Reader{r1, r2, r3}}
	}

	{
		cr := makeCR()
		bs, err := ioutil.ReadAll(cr)
		assert.NoError(err)
		s := string(bs)
		assert.EqualValues("foobarbaz", s)
	}

	{
		specs := [][]string{
			[]string{"foobar", "baz"},
			[]string{"f", "oobarb", "az"},
			[]string{"fo", "ob", "ar", "ba", "z"},
		}
		for _, spec := range specs {
			cr := makeCR()
			for _, f := range spec {
				bs := make([]byte, len(f))
				_, err := io.ReadAtLeast(cr, bs, len(f))
				assert.NoError(err)
				s := string(bs)
				assert.EqualValues(f, s)
			}
			_, err := cr.Read([]byte{0})
			assert.EqualValues(io.EOF, err)
		}
	}
}
