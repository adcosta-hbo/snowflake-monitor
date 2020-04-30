package secrets

type mockVaultClient struct {
	entries map[string]mockEntry
}

func (m *mockVaultClient) getObject(opts vaultStoreOpts, jwtLocation, key string) ([]byte, error) {
	return m.entries[key].byts, m.entries[key].err
}
