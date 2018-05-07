package main

import (
  "crypto/sha256"
       "encoding/hex"
       "encoding/json"
       "fmt"
       "io"
       "log"
       "net/http"
       "os"
       "strings"
       "sync"
       "time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

const difficulty = 1

//定义区块
type Block struct {
	Index     int
	Timestamp string
	Hash      string
	PrevHash  string
	IpfsHash  string
  Stats     int
  Dis       int
  Difficulty int
  Nonce     string
}

type Message struct {
	BPM        int
	IpfsHash   string
  Stats      int
  Dis        int
}

var Blockchain []Block
var mutex = &sync.Mutex{}


//计算区块的hash值
func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + block.IpfsHash + block.Nonce + string(block.Stats) + string(block.Dis) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}
//生成一个新的区块
func generateBlock(oldBlock Block, Stats int, Dis int, IpfsHash string) (Block) {

  var newBlock Block

       t := time.Now()

       newBlock.Index = oldBlock.Index + 1
       newBlock.Timestamp = t.String()
       newBlock.Stats = Stats
       newBlock.Dis = Dis
       newBlock.IpfsHash = IpfsHash
       newBlock.PrevHash = oldBlock.Hash
       newBlock.Difficulty = difficulty

       for i := 0; ; i++ {
               hex := fmt.Sprintf("%x", i)
               newBlock.Nonce = hex
               if !isHashValid(calculateHash(newBlock), newBlock.Difficulty) {
                       fmt.Println(calculateHash(newBlock), " do more work!")
                       time.Sleep(time.Second)
                       continue
               } else {
                       fmt.Println(calculateHash(newBlock), " work done!")
                       newBlock.Hash = calculateHash(newBlock)
                       break
               }

       }
       return newBlock
}

//验证所加的区块是否有效
func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

//用长链替换短链
func replaceChain(newBlocks []Block) {
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
}

//运行服务器
func run() error {
	mux := makeMuxRouter()
	httpAddr := os.Getenv("ADDR")
	log.Println("Listening on ", os.Getenv("ADDR"))
	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
//如果我们提交的是get请求，我们将会获得区块链，如果我们提交的是Post,我们将会写入区块链
func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}
func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(Blockchain, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

//post请求执行的回调
func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var m Message

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

  mutex.Lock()
	newBlock:= generateBlock(Blockchain[len(Blockchain)-1], m.Stats, m.Dis, m.IpfsHash)
  mutex.Unlock()

	/*if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, m)
		return
	}*/
	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		newBlockchain := append(Blockchain, newBlock)
		replaceChain(newBlockchain)
		spew.Dump(Blockchain)
	}

	respondWithJSON(w, r, http.StatusCreated, newBlock)

}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

func isHashValid(hash string, difficulty int) bool {
        prefix := strings.Repeat("0", difficulty)
        return strings.HasPrefix(hash, prefix)
}


func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		t := time.Now()
    genesisBlock := Block{}
		genesisBlock = Block{0, t.String(), calculateHash(genesisBlock), "", "", 0, 1,difficulty, ""}
		spew.Dump(genesisBlock)
    mutex.Lock()
		Blockchain = append(Blockchain, genesisBlock)
    mutex.Unlock()
	}()
	log.Fatal(run())

}
