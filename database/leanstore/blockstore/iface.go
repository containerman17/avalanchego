package blockstore

type IBlockPersister interface {
	Get(blockID uint32) ([]byte, error)
	Update(blockID uint32, data []byte) error
	Insert(data []byte) (uint32, error)
	Close() error
	NumBlocks() uint32
}
