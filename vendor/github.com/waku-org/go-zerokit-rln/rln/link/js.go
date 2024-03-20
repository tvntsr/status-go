// +build js


// ::FIXME:: it is not available in js
package link

import  "fmt"


type RLNWrapper struct {
}

func NewWithParams(depth int, wasm []byte, zkey []byte, verifKey []byte, treeConfig []byte) (*RLNWrapper, error) {
	return &RLNWrapper{}, nil
}

func New(depth int, config []byte) (*RLNWrapper, error) {
	return &RLNWrapper{}, nil
}

func (i RLNWrapper) SetTree(treeHeight uint) bool {
	return false
}

func (i RLNWrapper) InitTreeWithLeaves(idcommitments []byte) bool {
	return false
}

func (i RLNWrapper) KeyGen() []byte {
	return nil
}

func (i RLNWrapper) SeededKeyGen(seed []byte) []byte {
	return nil
}

func (i RLNWrapper) ExtendedKeyGen() []byte {
	return nil
}

func (i RLNWrapper) ExtendedSeededKeyGen(seed []byte) []byte {
	return nil
}

func (i RLNWrapper) Hash(input []byte) ([]byte, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (i RLNWrapper) PoseidonHash(input []byte) ([]byte, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (i RLNWrapper) SetLeaf(index uint, idcommitment []byte) bool {
	return false
}

func (i RLNWrapper) SetNextLeaf(idcommitment []byte) bool {
	return false
}

func (i RLNWrapper) SetLeavesFrom(index uint, idcommitments []byte) bool {
	return false
}

func (i RLNWrapper) DeleteLeaf(index uint) bool {
	return false
}

func (i RLNWrapper) GetRoot() ([]byte, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (i RLNWrapper) GetLeaf(index uint) ([]byte, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (i RLNWrapper) GenerateRLNProof(input []byte) ([]byte, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (i RLNWrapper) VerifyWithRoots(input []byte, roots []byte) (bool, error) {
	return false, fmt.Errorf("Not implemented")
}

func (i RLNWrapper) AtomicOperation(index uint, leaves []byte, indices []byte) bool {
	return false
}

func (i RLNWrapper) SeqAtomicOperation(leaves []byte, indices []byte) bool {
	return false
}

func (i RLNWrapper) RecoverIDSecret(proof1 []byte, proof2 []byte) ([]byte, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (i RLNWrapper) SetMetadata(metadata []byte) bool {
	return false
}

func (i RLNWrapper) GetMetadata() ([]byte, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (i RLNWrapper) Flush() bool {
	return false
}

func (i RLNWrapper) LeavesSet() uint {
	return 0
}
