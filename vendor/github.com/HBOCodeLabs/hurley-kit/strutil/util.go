package strutil

import "errors"

// Chunk splits a slice into one or more slices of size `chunkSize`. If the
// original slice isn't evenly divisible by `chunkSize`, the final chunk will
// contain the remaining elements.
func Chunk(slice []string, chunkSize int) ([][]string, error) {
	if chunkSize < 1 {
		return nil, errors.New("chunkSize must be >= 1")
	}

	if len(slice) == 0 {
		return [][]string{}, nil
	}

	if len(slice) <= chunkSize {
		return [][]string{slice}, nil
	}

	var (
		chunks    [][]string
		numChunks int
		remainder = len(slice) % chunkSize
	)

	if remainder == 0 {
		// the original slice is evenly divisible by 'chunkSize'
		numChunks = len(slice) / chunkSize
	} else {
		// allocate one extra slice to contain the remaining elements
		numChunks = (len(slice) / chunkSize) + 1
	}

	chunks = make([][]string, numChunks, numChunks)
	var chunk []string
	chunkIdx := 0

	for i, elem := range slice {
		if i%chunkSize == 0 {
			// time to allocate a new chunk

			// ensure that it's only as large as we need it to be: the
			// last chunk only needs to hold 'remainder' elements
			var chunkLen = chunkSize
			if remainder != 0 && chunkIdx+1 == numChunks {
				chunkLen = remainder
			}

			chunk = make([]string, chunkLen, chunkLen)
			chunks[chunkIdx] = chunk
			chunkIdx++
		}

		// insert the element into the current chunk
		chunk[i%chunkSize] = elem
	}

	return chunks, nil
}

// ElideString extracts `prefixLen` characters from the beginning, and
// `suffixLen` characters from the end of a string, and adds an ellipsis (...)
// in between. If either `prefixLen`, `suffixLen`, or the sum of the two is
// greater than the length of the string, the string is returned unmodified.
func ElideString(s string, prefixLen, suffixLen int) string {
	if len(s) < prefixLen || len(s) < suffixLen || len(s) < prefixLen+suffixLen {
		return s
	}

	if prefixLen < 1 && suffixLen < 1 {
		return "..."
	}

	if prefixLen < 1 {
		return "..." + s[len(s)-suffixLen:]
	}

	if suffixLen < 1 {
		return s[0:prefixLen] + "..."
	}

	return s[0:prefixLen] + "..." + s[len(s)-suffixLen:]
}
