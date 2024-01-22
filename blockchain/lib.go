package blockchain

import (
	"crypto/rand"
	"crypto/sha512"
	"tools"
)

type Hash [64]byte
type BlockData struct {
	Index        uint
	Nonce        []byte
	Data         []byte
	PreviousHash Hash
}

type Block struct {
	CurrentHash Hash
	BlockData   BlockData
}

type Blockchain struct {
	Chain      []Block
	difficulty []byte
}

func (b Block) calculateHash() Hash {
	a, err := tools.Encode(b.BlockData)
	tools.Panic(err)
	return sha512.Sum512(a)
}

func (b *Block) mine(difficulty []byte) {
	for {
		nonce := make([]byte, 8)
		rand.Read(nonce)
		b.BlockData.Nonce = nonce
		b.CurrentHash = b.calculateHash()
		if hasPrefix(b.CurrentHash, difficulty) {
			break
		}
	}
}

func CreateBlockchain(difficulty []byte) Blockchain {
	return Blockchain{
		[]Block{Block{}},
		difficulty,
	}
}

func (b *Blockchain) AddBlock(data []byte) {
	lastBlock := b.Chain[len(b.Chain)-1]
	newBlock := Block{
		CurrentHash: Hash{},
		BlockData: BlockData{
			Index:        lastBlock.BlockData.Index + 1,
			Nonce:        nil,
			Data:         data,
			PreviousHash: lastBlock.CurrentHash,
		},
	}
	newBlock.mine(b.difficulty)
	b.Chain = append(b.Chain, newBlock)
}

func (b Blockchain) IsValid() bool {
	for i := range b.Chain[1:] {
		previousBlock := b.Chain[i]
		currentBlock := b.Chain[i+1]
		if currentBlock.CurrentHash != currentBlock.calculateHash() || currentBlock.BlockData.PreviousHash != previousBlock.CurrentHash {
			return false
		}
	}
	return true
}

func (b Blockchain) IsValidPrefix() bool {
	for i := range b.Chain[1:] {
		currentBlock := b.Chain[i+1]
		if !hasPrefix(currentBlock.CurrentHash, b.difficulty) {
			return false
		}
	}
	return true
}

func hasPrefix(hash Hash, prefix []byte) bool {
	for i, v := range prefix {
		if v != hash[i] {
			return false
		}
	}
	return true
}
