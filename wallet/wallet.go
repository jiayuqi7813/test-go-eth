package wallet

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/howeyc/gopass"
	"github.com/tyler-smith/go-bip39"
	"log"
	"main.go/hdkeystore"
)

const defaultDerivationPath = "m/44'/60'/0'/0/1"

type HDwallet struct {
	Address common.Address
	HDKeyStore *hdkeystore.HDKeyStore
}

func NewWallet (keypath string)(*HDwallet,error){
	mn,err := Create_mnemonic()
	if err != nil {
		log.Fatal(err)
		return nil,err
	}
	//推导私钥
	privatekey,err := NewKeyFromMnemonic(mn)
	if err!=nil{
		log.Fatal(err)
		return nil,err
	}
	publicKey,err := DerivePublicKey(privatekey)
	if err!=nil{
		log.Fatal(err)
		return nil,err
	}
	//利用公钥推导私钥
	address := crypto.PubkeyToAddress(*publicKey)
	//创建keystore
	hdks := hdkeystore.NewHDKeyStore(keypath,privatekey)

	return &HDwallet{address,hdks},nil

}

func (w HDwallet)StoreKey(pass string)error{
	filename := w.HDKeyStore.JoinPath(w.Address.Hex())
	return w.HDKeyStore.StoreKey(filename,&w.HDKeyStore.Key,pass)
}


func Create_mnemonic() (string,error){
	//Entropy 生成，传入值y =32*x，且128<=y<=256
	b,err := bip39.NewEntropy(128)
	if err != nil {
		log.Fatal(err)
	}
	//生成助剂词
	nm,err := bip39.NewMnemonic(b)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(nm)
	return nm,nil

}

func LoadWallet(filename,datadir string) (*HDwallet,error){
	hdks := hdkeystore.NewHDkeyStoreNokey(datadir)
	fmt.Println("plase input password for :",filename)
	pass,_ :=gopass.GetPasswd()
	//filename为账户地址
	fromaddr := common.HexToAddress(filename)
	_,err := hdks.GetKey(fromaddr,hdks.JoinPath(filename),string(pass))
	if err != nil {
		log.Panic("failed to getkey",err)
	}
	return &HDwallet{fromaddr,hdks},nil

}

//推导私钥
func DerivePrivateKey(path accounts.DerivationPath,masterkey *hdkeychain.ExtendedKey)(*ecdsa.PrivateKey,error){
	var err error
	key := masterkey
	for _,index := range path {
		key,err = key.Child(index)
		if err != nil {
			return nil,err
		}
	}
	privKey,err := key.ECPrivKey()
	privateKeyECDSA := privKey.ToECDSA()
	if err !=nil{
		return nil,err
	}
	return privateKeyECDSA,nil
}

//推导公钥
func DerivePublicKey(privatekey *ecdsa.PrivateKey)(*ecdsa.PublicKey,error){
	publickKey :=privatekey.Public()
	publickKeyECDSA,OK :=publickKey.(*ecdsa.PublicKey)
	if !OK {
		return nil,fmt.Errorf("unknown public key type %T",publickKey)
	}
	return publickKeyECDSA,nil

}

func DeriveAddressFromMnemonic(){
	path ,err := accounts.ParseDerivationPath("m/44'/60'/0'/0/10")
	if err != nil {
		log.Fatal(err)
	}
	//获得种子
	nm:="tide track toe shy process stable pen antenna invite right priority evolve"
	seed,err := bip39.NewSeedWithErrorChecking(nm, "")
	//获得主key
	masetkey,err :=hdkeychain.NewMaster(seed,&chaincfg.MainNetParams)
	if err != nil {
		log.Fatal(err)

	}

	privatekey,err := DerivePrivateKey(path,masetkey)
	if err != nil {
		log.Fatal(err)
	}
	publickey,err := DerivePublicKey(privatekey)
	if err != nil {
		log.Fatal(err)
	}
	address := crypto.PubkeyToAddress(*publickey)
	fmt.Println(address.Hex())

}
func NewKeyFromMnemonic(mn string) (*ecdsa.PrivateKey, error) {
	//1. 先推导路径
	path, err := accounts.ParseDerivationPath(defaultDerivationPath)
	if err != nil {
		panic(err)
	}
	//2. 获得种子
	seed, err := bip39.NewSeedWithErrorChecking(mn, defaultDerivationPath)
	//3. 获得主key
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		fmt.Println("Failed to NewMaster", err)
		return nil, err
	}
	//4. 推导私钥
	privateKey, err := DerivePrivateKey(path, masterKey)

	return privateKey, nil
}
