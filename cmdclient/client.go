package client

import (
	"context"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"main.go/sol"
	"main.go/wallet"
	"math/big"
	"os"
)
type CmdClient struct {
	network string	//区块链地址
	dataDir string //数据路径
}

const TokenContractAddress ="0xACe30283322eF8E5Fb87ad14Fcd58228fB6a6016"

func (c CmdClient)Help(){
	fmt.Println("./main createwallet -pass PASSWORD --for create a new wallet")
	fmt.Println("./main transfer -from FROM -TOADDR TO -value value --for transfer value from FROM to TO")
	fmt.Println("./main balance -from FROM --for get balance of FROM")
	fmt.Println("./main sendtoken -from FROM -toaddr TOADDR -value VALUE --for sendtoken")
}



func NewCmdClient(network, datadir string) *CmdClient {
	return &CmdClient{network, datadir}
}

//f封装钱包
func (c CmdClient) CreateWallet(pass string) error{
	w,err := wallet.NewWallet(c.dataDir)
	if err != nil {
		return err
	}
	return w.StoreKey(pass)

}

//整合flag帮助方法



func (c CmdClient) Run(){
	if len(os.Args) <2{
		c.Help()
		os.Exit(-1)
	}
	cw_cmd := flag.NewFlagSet("createwallet",flag.ExitOnError)
	cw_cmd_pass := cw_cmd.String("pass","","passphrase for the new wallet")
	transfer_cmd := flag.NewFlagSet("transfer",flag.ExitOnError)
	transfer_cmd_from := transfer_cmd.String("from","","FROM")
	transfer_cmd_toaddr := transfer_cmd.String("toaddr","","TOADDR")
	transfer_cmd_value := transfer_cmd.Int64("value",0,"value")
	balance_cmd :=flag.NewFlagSet("balance",flag.ExitOnError)
	balance_cmd_from := balance_cmd.String("from","","FROM")
	tokenbalance_cmd := flag.NewFlagSet("tokenbalance",flag.ExitOnError)
	sendtoken_cmd_from := tokenbalance_cmd.String("from","","FROM")
	sendtoken_cmd_toaddr := tokenbalance_cmd.String("toaddr","","TOADDR")
	sendtoekn_cmd_value := tokenbalance_cmd.Int64("value",0,"VALUE")

	switch os.Args[1] {
	case "createwallet":
		err := cw_cmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("failed to parse cw_cmd",err)
			return
		}
		if cw_cmd.Parsed() {
			fmt.Println("params is ",*cw_cmd_pass)
			c.CreateWallet(*cw_cmd_pass)
		}
	case "transfer":
		err := transfer_cmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("failed to parse transfer_cmd",err)
			return
		}
		if transfer_cmd.Parsed() {
			fmt.Println(*transfer_cmd_from,*transfer_cmd_toaddr,*transfer_cmd_value)
			c.transfer(*transfer_cmd_from,*transfer_cmd_toaddr,*transfer_cmd_value)

		}
	case "balance":
		err := balance_cmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("failed to parse balance_cmd",err)
			return
		}
		if balance_cmd.Parsed(){
			c.balance(*balance_cmd_from)
		}
	case "sendtoken":
		err :=tokenbalance_cmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println("failed to parse tokenbalance_cmd",err)
			return
		}
		if tokenbalance_cmd.Parsed(){
			c.sendtoken(*sendtoken_cmd_from,*sendtoken_cmd_toaddr,*sendtoekn_cmd_value)
		}


	}




}

//coin转移

func (c CmdClient) transfer(from, toaddr string ,value int64) error{
	//钱包加载
	w,_ := wallet.LoadWallet(from,c.dataDir)
	//连接到以太坊
	cli,_ :=ethclient.Dial(c.network)
	defer cli.Close()
	//获取nonce
	nonce,_ :=cli.NonceAt(context.Background(),common.HexToAddress(from),nil)
	//创建交易
	gaslimit := uint64(300000)
	gasprice := big.NewInt(21000000000)
	amount :=big.NewInt(value)
	tx := types.NewTransaction(nonce,common.HexToAddress(toaddr),amount,gaslimit,gasprice, []byte("Salary"))
	//签名交易
	stx,err := w.HDKeyStore.SignTx(common.HexToAddress(from),tx,nil)
	if err != nil {
		log.Panic("failed to sign tx",err)
	}
	//发送交易
	err = cli.SendTransaction(context.Background(),stx)
	return err

}

func (c CmdClient)balance(from string) (int64,error){
	cli , err := ethclient.Dial(c.network)
	if err != nil {
		log.Panic("failed to dial",err)
	}
	defer cli.Close()
	addr := common.HexToAddress(from)
	value,err :=cli.BalanceAt(context.Background(),addr,nil)
	if err != nil {
		log.Panic("failed to get balance",err)
	}
	fmt.Printf("%s's balance is %d\n",from,value)
	return value.Int64(),nil
}


func (c CmdClient) sendtoken(from,toaddr string,value int64)error{
	//连接以太坊
	cli,err := ethclient.Dial(c.network)
	if err != nil {
		log.Panic("failed to dial",err)
	}
	defer cli.Close()
	//设置合约地址
	token,_:=sol.NewToken(common.HexToAddress(TokenContractAddress),cli)
	w,_:=wallet.LoadWallet(from,c.dataDir)
	auth := w.HDKeyStore.NewTransactOpts()
	_,err = token.Transfer(auth,common.HexToAddress(toaddr),big.NewInt(value))
	return err

}