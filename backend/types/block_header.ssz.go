package types

import (
	"github.com/ferranbt/fastssz"
)

// MarshalSSZ ssz marshals the BlockHeader object
func (b *BlockHeader) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(b)
}

// MarshalSSZTo ssz marshals the BlockHeader object to a target array
func (b *BlockHeader) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf

	// Field (0) 'Slot'
	dst = ssz.MarshalUint64(dst, uint64(b.Slot))

	// Field (1) 'ProposerIndex'
	dst = ssz.MarshalUint64(dst, uint64(b.ProposerIndex))

	// Field (2) 'ParentRoot'
	if size := len(b.ParentRoot); size != 32 {
		err = ssz.ErrBytesLengthFn("--.ParentRoot", size, 32)
		return
	}
	dst = append(dst, b.ParentRoot...)

	// Field (3) 'StateRoot'
	if size := len(b.StateRoot); size != 32 {
		err = ssz.ErrBytesLengthFn("--.StateRoot", size, 32)
		return
	}
	dst = append(dst, b.StateRoot...)

	// Field (4) 'BodyRoot'
	if size := len(b.BodyRoot); size != 32 {
		err = ssz.ErrBytesLengthFn("--.BodyRoot", size, 32)
		return
	}
	dst = append(dst, b.BodyRoot...)

	return
}

// UnmarshalSSZ ssz unmarshals the BlockHeader object
func (b *BlockHeader) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size != 112 {
		return ssz.ErrSize
	}

	// Field (0) 'Slot'
	b.Slot = uint64(ssz.UnmarshallUint64(buf[0:8]))

	// Field (1) 'ProposerIndex'
	b.ProposerIndex = uint64(ssz.UnmarshallUint64(buf[8:16]))

	// Field (2) 'ParentRoot'
	if cap(b.ParentRoot) == 0 {
		b.ParentRoot = make([]byte, 0, len(buf[16:48]))
	}
	b.ParentRoot = append(b.ParentRoot, buf[16:48]...)

	// Field (3) 'StateRoot'
	if cap(b.StateRoot) == 0 {
		b.StateRoot = make([]byte, 0, len(buf[48:80]))
	}
	b.StateRoot = append(b.StateRoot, buf[48:80]...)

	// Field (4) 'BodyRoot'
	if cap(b.BodyRoot) == 0 {
		b.BodyRoot = make([]byte, 0, len(buf[80:112]))
	}
	b.BodyRoot = append(b.BodyRoot, buf[80:112]...)

	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the BlockHeader object
func (b *BlockHeader) SizeSSZ() (size int) {
	size = 112
	return
}

// HashTreeRoot ssz hashes the BlockHeader object
func (b *BlockHeader) HashTreeRoot() ([32]byte, error) {
	hh := ssz.DefaultHasherPool.Get()
	defer ssz.DefaultHasherPool.Put(hh)
	err := b.HashTreeRootWith(hh)
	if err != nil {
		return [32]byte{}, err
	}
	return hh.HashRoot()
}

// HashTreeRootWith ssz hashes the BlockHeader object with a hasher
func (b *BlockHeader) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Slot'
	hh.PutUint64(uint64(b.Slot))

	// Field (1) 'ProposerIndex'
	hh.PutUint64(uint64(b.ProposerIndex))

	// Field (2) 'ParentRoot'
	if size := len(b.ParentRoot); size != 32 {
		err = ssz.ErrBytesLengthFn("--.ParentRoot", size, 32)
		return
	}
	hh.PutBytes(b.ParentRoot)

	// Field (3) 'StateRoot'
	if size := len(b.StateRoot); size != 32 {
		err = ssz.ErrBytesLengthFn("--.StateRoot", size, 32)
		return
	}
	hh.PutBytes(b.StateRoot)

	// Field (4) 'BodyRoot'
	if size := len(b.BodyRoot); size != 32 {
		err = ssz.ErrBytesLengthFn("--.BodyRoot", size, 32)
		return
	}
	hh.PutBytes(b.BodyRoot)

	hh.Merkleize(indx)
	return
}