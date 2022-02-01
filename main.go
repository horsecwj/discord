package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/fsnotify/fsnotify"
	"github.com/gocarina/gocsv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/viper"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func Config() *viper.Viper {
	onceConfig.Do(func() {
		cfgInit()
	})
	return config
}

var (
	config     *viper.Viper
	onceConfig sync.Once
)

func init() {
	pwd, _ := os.Getwd()
	fmt.Println("开始工作目录", pwd)
	// 程序所在目录
	execDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	if pwd == execDir {
		fmt.Println("不需要切换工作目录")
		return
	}
	fmt.Println("切换工作目录到", execDir)
	if err := os.Chdir(execDir); err != nil {
		log.Fatal(err)
	}
	pwd, _ = os.Getwd()
	fmt.Println("切换后工作目录:", pwd)
}

//读取配置文件
func cfgInit() {

	config = viper.New()
	config.SetConfigName("config")
	config.SetConfigType("yaml")
	config.AddConfigPath("./")
	if err := config.ReadInConfig(); err != nil {
		if err != nil {
			log.Fatalf("Fatal error config file: %s \n", err)
		}
	}
	config.OnConfigChange(func(e fsnotify.Event) {
		if errs := config.ReadInConfig(); errs != nil {
			if errs != nil {
				log.Printf("Fatal error config file: %s \n", errs)
			}
		}
		log.Printf("config is change :%s \n", e.String())
	})
	//开始监听
	go config.WatchConfig()

}

func botInfo() (string, string, string) {
	appId := Config().GetString("AppId")
	secretId := Config().GetString("secretId")
	GuildId := Config().GetString("GuildId")
	server := Config().GetString("Server")
	token := Config().GetString("Token")
	msg := Config().GetString("Msg")

	log.Printf("appId : %s \n secretId :%s   \n server : %s  \n   token :%s  \n msg %s ", appId, secretId, server, token, msg)
	return token, GuildId, msg
}

func main() {
	discordBot()
}

func discordBot() {

	token, Gid, msg := botInfo()
	ForNewUIdMsg = msg
	// creates a new Discord session
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	KeyMap := LoadUid()

	GetNotBotMembers(dg, Gid, KeyMap)
	reLoadFile(AllUId, "AllUId")
	log.Printf("总待发送 %d ", len(AllUId))

	SendUId(dg)
	reLoadFile(SendSuccessUId, "SendSuccessUId")
	log.Printf("成功发送 %d", len(SendSuccessUId))

	dg.AddHandler(messageAdd)

	err = dg.Open()
	defer dg.Close()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	fmt.Println("Bot is now running.")
	for {
		select {}
	}

}

type StrObj struct {
	ID string
}

var AllUId []StrObj
var SendSuccessUId []StrObj
var ForNewUIdMsg string

func SendUId(dg *discordgo.Session) {
	for _, item := range AllUId {
		re, err := dg.UserChannelCreate(item.ID)
		if err != nil {
			log.Print(err)
		}
		_, err = dg.ChannelMessageSend(re.ID, ForNewUIdMsg)
		if err != nil {
			log.Print(err)
		} else {
			SendSuccessUId = append(SendSuccessUId, item)
		}
		s := rand.Intn(2-1) + 1
		fmt.Println(s)
		time.Sleep(time.Second * time.Duration(s))

	}
}

func GetNotBotMembers(dg *discordgo.Session, gId string, keyMap map[string]bool) {
	startId := "0"

	for i := 0; i <= 1000; i++ {
		re, err := dg.GuildMembers(gId, startId, 1000)
		if err != nil {
			return
		}
		if len(re) == 0 {
			break
		}
		startId = re[len(re)-1].User.ID
		for _, item := range re {
			if !item.User.Bot && !keyMap[item.User.ID] {
				AllUId = append(AllUId, StrObj{ID: item.User.ID})
			}
		}

		s := rand.Intn(2-1) + 1
		fmt.Println(s)
		time.Sleep(time.Second * time.Duration(s))

	}
}

func SendMsgById(dg *discordgo.Session, id, msg string) error {
	st, err := dg.UserChannelCreate(id)
	if err != nil {
		return err
	}
	_, err = dg.ChannelMessageSend(st.ID, msg)
	if err != nil {
		return err
	}
	return err
}

func messageAdd(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Type == discordgo.MessageType(7) {
		err := SendMsgById(s, m.Author.ID, ForNewUIdMsg)
		if err != nil {
			log.Print(err)
		} else {
			SendSuccessUId = append(SendSuccessUId, StrObj{ID: m.Author.ID})
			reLoadFile(SendSuccessUId, "SendSuccessUId")
			log.Printf("成功发送%d", len(SendSuccessUId))
		}
	}
}

func reLoadFile(finalBackTest []StrObj, str string) {

	clientsFile, err := os.OpenFile("./"+str+".csv", os.O_RDWR|os.O_CREATE, os.ModePerm)

	if err != nil {
		log.Print(err)
	}
	err = gocsv.MarshalFile(&(finalBackTest), clientsFile) // Use this to save the CSV back to the file
	if err != nil {
		log.Print(err)
	}
	clientsFile.Close()
	return
}

func LoadUid() map[string]bool {
	var UidKey map[string]bool
	UidKey = make(map[string]bool)
	var finalBackTest []StrObj
	clientsFile, err := os.OpenFile("./"+"SendSuccessUId"+".csv", os.O_RDWR|os.O_CREATE, os.ModePerm)

	if err != nil {
		log.Print(err)
	}
	err = gocsv.Unmarshal(clientsFile, &finalBackTest) // Use this to save the CSV back to the file
	if err != nil {
		log.Print(err)
	}
	for _, item := range finalBackTest {
		UidKey[item.ID] = true
	}
	clientsFile.Close()
	return UidKey
}
