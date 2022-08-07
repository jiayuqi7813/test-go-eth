package hdkeystore

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	uuid2 "github.com/google/uuid"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type HDKeyStore struct {
	keysDirPath string	//文件所在路径
	scryptN int			//生成加密文件的参数N
	scryptP int			//生成加密文件的参数P
	Key 	keystore.Key	//keystroe对应的key

}
func NewHDkeyStoreNokey(path string) *HDKeyStore{
	return &HDKeyStore{
		keysDirPath: path,
		scryptN: keystore.LightScryptN,
		scryptP: keystore.LightScryptP,
		Key: keystore.Key{},
	}
}


type UUID []byte
var rander = rand.Reader



func NewHDKeyStore(path string,privateKey *ecdsa.PrivateKey) *HDKeyStore{
	uuid :=uuid.NewV4()
	key := keystore.Key{
		Id: uuid2.UUID(uuid),
		Address:    crypto.PubkeyToAddress(privateKey.PublicKey),
		PrivateKey: privateKey,

	}
	return &HDKeyStore{
		keysDirPath: path,
		scryptN: keystore.LightScryptN,
		scryptP: keystore.LightScryptP,
		Key: key,
	}
}


//存储key为keystore文件
func (ks HDKeyStore) StoreKey(filename string,key *keystore.Key,auth string)error{
	keyjson,err :=keystore.EncryptKey(key,auth,ks.scryptN,ks.scryptP)
	if err != nil {
		return err
	}
	//写入文件
	return WriteKeyFile(filename,keyjson)

}

//写入key文件
func WriteKeyFile(file string,content []byte) error{
	const dirprem = 0700
	if err := os.MkdirAll(filepath.Dir(file),dirprem);err != nil {
		return err
	}
	f, err := ioutil.TempFile(filepath.Dir(file), "."+filepath.Base(file)+".tmp")
	if err != nil {
		return err
	}
	if _,err := f.Write(content);err !=nil{
		f.Close()
		os.Remove(f.Name())
		return err
	}
	f.Close()
	return os.Rename(f.Name(),file)
}

func (ks HDKeyStore) JoinPath(filename string) string{
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(ks.keysDirPath,filename)
}

//解析key
func (ks *HDKeyStore) GetKey(addr common.Address,filename,auth string)(*keystore.Key,error){
	keyjson,err := ioutil.ReadFile(filename)
	if err != nil {
		return nil,err
	}
	key,err :=keystore.DecryptKey(keyjson,auth)
	if err != nil {
		return nil,err
	}
	if key.Address!=addr{
		return nil,fmt.Errorf("key content mismatch: have address %x, want %x",key.Address,addr)
	}
	ks.Key = *key
	return key,nil
}

func (ks *HDKeyStore) SignTx(address common.Address, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {

	// Sign the transaction and verify the sender to avoid hardware fault surprises
	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, ks.Key.PrivateKey)
	if err != nil {
		return nil, err
	}

	//验证 remix->msg
	msg, err := signedTx.AsMessage(types.HomesteadSigner{}, chainID)

	if err != nil {
		return nil, err
	}

	sender := msg.From()
	if sender != address {
		return nil, fmt.Errorf("signer mismatch: expected %s, got %s", address.Hex(), sender.Hex())
	}

	return signedTx, nil
}


func (ks HDKeyStore) NewTransactOpts() *bind.TransactOpts{
	return bind.NewKeyedTransactor(ks.Key.PrivateKey)
}