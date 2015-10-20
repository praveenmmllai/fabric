/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package openchain

import (
	"fmt"
	"github.com/openblockchain/obc-peer/protos"
	"golang.org/x/net/context"
	"sync"
)

// Ledger - the struct for openchain ledger
type Ledger struct {
	blockchain *Blockchain
	state      *State
	currentID  interface{}
}

var ledger *Ledger
var mutex sync.Mutex

// GetLedger - gives a reference to a singleton ledger
func GetLedger() (*Ledger, error) {
	if ledger == nil {
		mutex.Lock()
		defer mutex.Unlock()
		if ledger == nil {
			blockchain, err := GetBlockchain()
			if err != nil {
				return nil, err
			}
			state := GetState()
			ledger = &Ledger{blockchain, state, nil}
		}
	}
	return ledger, nil
}

// BeginTxBatch - gets invoked when next round of transaction-batch execution begins
func (ledger *Ledger) BeginTxBatch(id interface{}) error {
	err := ledger.checkValidID(id)
	if err != nil {
		return err
	}
	ledger.currentID = id
	return nil
}

// CommitTxBatch - gets invoked when the current transaction-batch needs to be committed
// This function returns successfully iff the transactions details and state changes (that
// may have happened during execution of this transaction-batch) have been committed to permanent storage
func (ledger *Ledger) CommitTxBatch(id interface{}, transactions []*protos.Transaction, proof []byte) error {
	err := ledger.checkValidID(id)
	if err != nil {
		return err
	}
	block := protos.NewBlock("proposerID string", transactions)
	ledger.blockchain.AddBlock(context.TODO(), block)
	ledger.resetForNextTxGroup()
	return nil
}

// RollbackTxBatch - Descards all the state changes that may have taken place during the execution of
// current transaction-batch
func (ledger *Ledger) RollbackTxBatch(id interface{}) error {
	err := ledger.checkValidID(id)
	if err != nil {
		return err
	}
	ledger.resetForNextTxGroup()
	return nil
}

// GetTempStateHash - Computes state hash by taking into account the state changes that may have taken
// place during the execution of current transaction-batch
func (ledger *Ledger) GetTempStateHash() ([]byte, error) {
	return ledger.state.GetTempStateHash()
}

func (ledger *Ledger) checkValidID(id interface{}) error {
	if ledger.currentID != nil {
		return fmt.Errorf("Another TxGroup [%s] already in-progress", ledger.currentID)
	}
	return nil
}

func (ledger *Ledger) resetForNextTxGroup() {
	ledger.currentID = nil
	ledger.state.ClearInMemoryChanges()
}
