package main

import (
    "bufio"
  	"crypto/sha256"
  	"encoding/hex"
  	"encoding/json"
  	"io"
  	"log"
  	"net"
  	"os"
  	"strconv"
  	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)

//定义区块
type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
  Stats     int
  Dis       int
}

type Message struct {
	BPM        int
  Stats      int
  Dis        int
}

var Blockchain []Block
var bcServer chan []Block
//计算区块的hash值
func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.BPM) + string(block.Stats) + string(block.Dis) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}
//生成一个新的区块
func generateBlock(oldBlock Block, Stats int, Dis int) (Block, error) {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	//newBlock.BPM = BPM
  newBlock.Stats = Stats
  newBlock.Dis = Dis
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
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


func handleConn(conn net.Conn) {
	defer conn.Close()



}


func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

  bcServer = make(chan []Block)


  t := time.Now()
	genesisBlock := Block{0, t.String(), 0, "", "", 0, 1}
	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock)

  // start TCP and serve TCP server
  	server, err := net.Listen("tcp", ":"+os.Getenv("ADDR"))
  	if err != nil {
  		log.Fatal(err)
  	}
  	defer server.Close()

    for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)


    io.WriteString(conn, "Enter a new Stats:")
    //io.WriteString(conn, "Enter a new Dis:")
  	scanner := bufio.NewScanner(conn)

  	// take in BPM from stdin and add it to blockchain after conducting necessary validation
  	go func() {
  		for scanner.Scan() {
  			stats, err := strconv.Atoi(scanner.Text())
        //dis, err := strconv.Atoi(scanner.Text())
        dis := 0
  			if err != nil {
  				log.Printf("%v not a number: %v", scanner.Text(), err)
  				continue
  			}
  			newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], stats, dis)
  			if err != nil {
  				log.Println(err)
  				continue
  			}
  			if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
  				newBlockchain := append(Blockchain, newBlock)
  				replaceChain(newBlockchain)
  			}

  			bcServer <- Blockchain
  			io.WriteString(conn, "\nEnter a new Stats:")
  		}
  	}()


    // simulate receiving broadcast
  	go func() {
  		for {
  			time.Sleep(30 * time.Second)
  			output, err := json.Marshal(Blockchain)
  			if err != nil {
  				log.Fatal(err)
  			}
  			io.WriteString(conn, string(output))
  		}
  	}()

  	for _ = range bcServer {
  		spew.Dump(Blockchain)
  	}


	}



}
