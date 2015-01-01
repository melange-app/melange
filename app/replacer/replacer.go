package replacer

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"strings"
)

type Replacer struct {
	// Underlying Reader
	io.Reader

	// Internal Buffers
	unprocessedData *bytes.Buffer
	processedData   *bytes.Buffer

	// Internal Regular Expressions
	replaceString string
	literalPrefix string
	checkString   *regexp.Regexp
	stopChars     *regexp.Regexp

	testerLength int
}

// Example: we want a writer that will replace
// http://(*).melange(:7776) --> http://$1.melange.127.0.0.1.xip.io:7776
// while the stream is being written, we do this by buffering the input.

// http://common.melange
// http://app.melange
// http://api.melange
// http://(*).plugins.melange

const SMALL_READ = 65535

func CreateReplacer(r io.Reader, start, end, stopChars string) *Replacer {
	// Compile what we are replacing
	exp := regexp.MustCompile(start)
	prefix, _ := exp.LiteralPrefix()

	// Compile Stop Characters
	stopExp := regexp.MustCompile(stopChars)

	testerLength := 2 * len(prefix)
	if testerLength < SMALL_READ {
		testerLength = SMALL_READ
	}

	return &Replacer{
		// Underlying Reader
		Reader: r,

		// Initialize Buffers
		unprocessedData: new(bytes.Buffer),
		processedData:   new(bytes.Buffer),

		// Initialize Expressions
		replaceString: end,
		checkString:   exp,
		literalPrefix: prefix,
		stopChars:     stopExp,

		// Initialize Parameters
		testerLength: testerLength,
	}
}

// Prefix: http://
// Stop Characters: [^a-z\.]
// strings.Index(prefix)
// regexp.FindIndex(stopChar)

// Strategy:
// (1) Find literal prefix
// (2) Munch until we hit a stop-character
// (3) Check match
// (4) [Possibly,] Replace!

var readerNoMatch error = errors.New("getmelange.com/app.(*Replacer) Couldn't read off enough unprocessed data.")
var writerNoMatch error = errors.New("getmelange.com/app.(*Replacer) Couldn't write enough processed data.")

func (r *Replacer) Close() error {
	if closer, ok := r.Reader.(io.ReadCloser); ok {
		return closer.Close()
	}

	return errors.New("Couldn't close a non-closer.")
}

func (r *Replacer) Read(data []byte) (int, error) {
	// If we have enough processed data, just return that
	if r.processedData.Len() >= len(data) {
		return r.processedData.Read(data)
	}

	// We always pull in two times the length of the prefix to ensure
	// that we catch it
	testerLength := r.testerLength

	// Loop until we have enough in processed data to return
	for r.processedData.Len() < len(data) {
		// Create a reader that will start with unprocessedData, then
		// move to the underlying reader
		multiReader := io.MultiReader(r.unprocessedData, r.Reader)

		tester := make([]byte, testerLength)

		n, err := io.ReadAtLeast(multiReader, tester, len(tester))

		if err == io.ErrUnexpectedEOF || n < len(tester) {
			// Replace all that we have left
			replaced := r.checkString.ReplaceAll(tester[:n], []byte(r.replaceString))

			// Write it into processedData
			n, err := r.processedData.Write(replaced)
			if err != nil {
				return -1, err
			} else if n != len(replaced) {
				return -1, writerNoMatch
			}

			return r.processedData.Read(data)
		} else if err != nil {
			return -1, err
		} else if n != len(tester) {
			return -1, readerNoMatch
		}

		prefixIndex := strings.Index(string(tester), r.literalPrefix)
		if prefixIndex != -1 {
			intBuffer := new(bytes.Buffer)
			for {
				intermediary := make([]byte, SMALL_READ)

				n, err := multiReader.Read(intermediary)

				// Write all that we have to the intermediary buffer
				intN, intErr := intBuffer.Write(intermediary[:n])
				if intErr != nil {
					return -1, intErr
				} else if intN != n {
					return -1, writerNoMatch
				}

				// Construct bytes that represent everything that we are looking in
				// for matches
				restOfTester := append(tester[prefixIndex:], intBuffer.Bytes()...)
				if err == io.EOF || n < len(intermediary) {
					// Replace all that we have left
					replaced := r.checkString.ReplaceAll(restOfTester, []byte(r.replaceString))

					n, err := r.processedData.Write(tester[:prefixIndex])
					if err != nil {
						return -1, err
					} else if n != prefixIndex {
						return -1, writerNoMatch
					}

					// Write it into processedData
					n, err = r.processedData.Write(replaced)
					if err != nil {
						return -1, err
					} else if n != len(replaced) {
						return -1, writerNoMatch
					}

					// Aww, break out
					return r.processedData.Read(data)
				} else if err != nil {
					return -1, err
				} else if n != len(intermediary) {
					return -1, readerNoMatch
				}

				locations := r.stopChars.FindIndex(restOfTester)
				if locations != nil {
					// If we have a match, we should replace everything between the
					// prefix and the stop characters.
					n, err := r.processedData.Write(tester[:prefixIndex])
					if err != nil {
						return -1, err
					} else if n != prefixIndex {
						return -1, writerNoMatch
					}

					replacementBytes := r.checkString.ReplaceAll(
						append(tester[prefixIndex:], intBuffer.Bytes()[:locations[0]]...),
						[]byte(r.replaceString))
					n, err = r.processedData.Write(replacementBytes)
					if err != nil {
						return -1, err
					} else if n != len(replacementBytes) {
						return -1, writerNoMatch
					}

					unprocessedData := intBuffer.Bytes()[locations[0]:]
					n, err = r.unprocessedData.Write(unprocessedData)
					if err != nil {
						return -1, err
					} else if n != len(unprocessedData) {
						return -1, writerNoMatch
					}
				}

				// If there are no matches, we will just keep reading...
			}

		} else {
			// Can't find the prefix?
			// Just write it to processed data...
			halfOfTester := testerLength / 2
			n, err := r.processedData.Write(tester[:halfOfTester])
			if err != nil {
				return -1, err
			} else if n != halfOfTester {
				return -1, writerNoMatch
			}

			n, err = r.unprocessedData.Write(tester[halfOfTester:])
			if err != nil {
				return -1, err
			} else if n != halfOfTester {
				return -1, writerNoMatch
			}
		}
	}

	return r.processedData.Read(data)
}
