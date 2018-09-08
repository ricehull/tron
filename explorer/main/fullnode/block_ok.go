package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/tronprotocol/grpc-gateway/core"
	"github.com/wlcy/tron/explorer/core/grpcclient"
	"github.com/wlcy/tron/explorer/core/utils"
)

var bulkFetchLimit = int64(100)
var maxErrCnt = 60

func getBlock(id int, b, e int64) {
	startWorker()

	ts := time.Now()

	servAddr := fmt.Sprintf("%s:50051", utils.GetRandFullNode())
	taskID := fmt.Sprintf("[%04v|%v~%v|%v]", id, b, e, servAddr)

	client := grpcclient.NewWallet(servAddr)
	client.Connect()
	dbc := grpcclient.NewDatabase(servAddr)
	dbc.Connect()

	le := getLatestNum(dbc)
	if le == 0 {
		stopWorker()
		getBlock(id, b, e)
		return
	}
	fmt.Printf("%v latestNum is [%v]\n", taskID, le)
	b = checkForkTask(id, "", le, b, e)

	bb := b
	cnt := int64(0)
	errCnt := 0

	for {

		if errCnt >= maxErrCnt {
			stopWorker()
			getBlock(id, b, e) // redo full bulk of block
			return
		}

		if e > 0 && b >= e {
			break
		}

		if id == 0 && b >= le {
			time.Sleep(3 * time.Second)

			le = getLatestNum(dbc)
			runTaskCnt := workingTaskCnt()
			fmt.Printf("Current working task:[%v]--max task:[%v]\n", runTaskCnt, *gIntMaxWorker)
			if e > 0 && 1 == runTaskCnt {
				fmt.Printf("Sync all data cost:%v\n", time.Since(ts))
				break
			}
		}

		newE := b + bulkFetchLimit

		if e > 0 && newE > e {
			newE = e
		} else if e == 0 && newE > le {
			newE = le
		}

		blocks, err := client.GetBlockByLimitNext(b, newE)
		if nil != err {
			errCnt++
		}

		ret := verifyStoreBlock(blocks, genVerifyBlockIDList(b, newE), client, maxErrCnt-errCnt)
		if !ret {
			fmt.Printf("bulk get block(%v, %v) check store failed! error:%v\n", b, newE, err)
			errCnt += maxErrCnt
		}

		c := int64(len(blocks))
		cnt += c
		b += c
	}
	fmt.Printf("%v Finish work, total cost:%v, total block:%v(%v), begin:%v, end:%v\n", taskID, time.Since(ts), cnt, b-bb, bb, b)

	stopWorker()
}

func getBlockByIDs(blockIDs []int64, client *grpcclient.Wallet) ([]*core.Block, []int64) {
	ret := make([]*core.Block, 0, len(blockIDs))
	missingBlockID := make([]int64, 0)
	for _, id := range blockIDs {
		block, err := client.GetBlockByNum(id)
		if err == nil && nil != block && nil != block.BlockHeader && nil != block.BlockHeader.RawData && block.BlockHeader.RawData.Number == id {
			ret = append(ret, block)
		} else {
			missingBlockID = append(missingBlockID, id)
		}
	}

	return ret, missingBlockID
}

func getLatestNum(dbc *grpcclient.Database) int64 {
	prop, err := dbc.GetDynamicProperties()
	if nil == err && nil != prop {
		return prop.LastSolidityBlockNum
	}
	return 0
}

func checkForkTask(id int, taskID string, latestE, b, e int64) (newB int64) {
	newB = b
	if e == 0 {
		if id != 0 { // e == 0 only for task id == 0
			return
		}

		if latestE-b > *gInt64MaxWorkload {
			newB = latestE - *gInt64MaxWorkload
			forkBlockTask(id+1, b, newB)
		}
	} else {
		if e-b > *gInt64MaxWorkload {
			newB = e - *gInt64MaxWorkload
			forkBlockTask(id+1, b, newB)
		}
	}
	return
}

func forkBlockTask(id int, b, e int64) {
	go getBlock(id, b, e)
}

func genVerifyBlockIDList(b, e int64) (ret []int64) {
	for i := b; i < e; i++ {
		ret = append(ret, i)
	}
	return
}

func verifyStoreBlock(blocks []*core.Block, blockIDCheckList []int64, client *grpcclient.Wallet, retryCnt int) bool {
	if len(blocks) == 0 {
		return true
	}
	_, succCnt, errCnt, blockList := storeBlocks(blocks)

	sort.Slice(blockList, func(i, j int) bool { return blockList[i] < blockList[j] })

	missingBlockID := make([]int64, 0)
	blockCnt := len(blockList)
	for _, blockID := range blockIDCheckList {
		retIdx := sort.Search(blockCnt, func(idx int) bool { return blockList[idx] >= blockID })

		if retIdx < blockCnt && blockList[retIdx] == blockID {

		} else {
			missingBlockID = append(missingBlockID, blockID)
		}
	}
	if len(missingBlockID) > 0 {
		fmt.Printf("missing %v, try cnt remain:%v raw block size:%v, succ:%v, err:%v \n\tmissing blockIDs(%v):%v\n\tpull blockIDs(%v):%v\n", blockIDCheckList, retryCnt, len(blocks), succCnt, errCnt, len(missingBlockID), missingBlockID, len(blockList), blockList)

		if retryCnt == 0 {
			return false
		}

		blocks, _ := getBlockByIDs(missingBlockID, client)

		return verifyStoreBlock(blocks, missingBlockID, client, retryCnt-1)
	}

	return true

}
