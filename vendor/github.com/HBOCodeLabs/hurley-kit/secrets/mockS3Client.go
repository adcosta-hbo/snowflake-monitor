package secrets

type mockS3Client struct {
	entries map[string]mockEntry
}

func (m *mockS3Client) getObject(region, bucket, key string) ([]byte, error) {
	return m.entries[key].byts, m.entries[key].err
}
