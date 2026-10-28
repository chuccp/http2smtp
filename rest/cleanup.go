package rest

// deleteInBatches runs deleteFn over ids in chunks. A single "id IN (...)" statement
// binds one variable per id, and databases cap that (SQLite allows 32766 in modern
// builds, 999 in old ones); the log tables are the ones that can grow past it. Chunking
// also lets the caller clean up after each batch, so a failure halfway leaves the rows it
// already removed consistent.
func deleteInBatches(ids []uint, deleteFn func(batch []uint) error) error {
	const batchSize = 500
	for start := 0; start < len(ids); start += batchSize {
		end := min(start+batchSize, len(ids))
		if err := deleteFn(ids[start:end]); err != nil {
			return err
		}
	}
	return nil
}
