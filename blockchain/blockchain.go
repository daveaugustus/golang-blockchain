package blockchain

import (
	"fmt"

	"github.com/dgraph-io/badger/v3"
)

const (
	dbPath = "./tmp/blocks"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func InitBlockChain() *BlockChain {
	var lastHash []byte

	opts := badger.DefaultOptions(dbPath) // Use DefaultOptions for simplicity
	db, err := badger.Open(opts)
	if err != nil {
		fmt.Printf("Error opening BadgerDB: %v\n", err)
		return nil
	}

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found")
			genesis := Genesis() // You need to implement the Genesis function
			fmt.Println("Genesis proved")

			err = txn.Set(genesis.Hash, genesis.Serialize())
			if err != nil {
				return err
			}

			err = txn.Set([]byte("lh"), genesis.Hash)
			if err != nil {
				return err
			}

			lastHash = genesis.Hash
			return nil
		} else if err != nil {
			return err
		}

		lastHash, err = item.ValueCopy(nil) // Use ValueCopy to get the value
		return err
	})

	if err != nil {
		fmt.Printf("Error initializing blockchain: %v\n", err)
		return nil
	}

	blockchain := BlockChain{LastHash: lastHash, Database: db}
	return &blockchain
}

func (chain *BlockChain) AddBlock(data string) {
	var lastHash []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)

		lastHash, err = item.ValueCopy(nil)
		return err
	})
	Handle(err)

	newBlock := CreateBlock(data, lastHash)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	})
	Handle(err)
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{CurrentHash: chain.LastHash, Database: chain.Database}
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)

		encodedBlock, err := item.ValueCopy(nil)
		block = Deserialize(encodedBlock)
		return err
	})
	Handle(err)

	iter.CurrentHash = block.PrevHash

	return block
}
