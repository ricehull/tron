/**
 * @author [yanzheng]
 * @email [yanzheng@mail.com]
 * @create date 2018-06-21 03:25:57
 * @modify date 2018-06-21 03:25:57
 * @desc [服务初始化加载项]
 */

package config

import (
	"fmt"

	"github.com/wlcy/tron/explorer/lib/log"
	"github.com/wlcy/tron/explorer/lib/mysql"
	"github.com/wlcy/tron/explorer/lib/redis"
	"github.com/wlcy/tron/explorer/lib/util"

	"github.com/pelletier/go-toml"
)

//配置信息
var RedisCli *redis.TronRedis

var DefaultPath string
var TokenTemplate string
var ImgURL string
var Address string
var TokenTemplateFile string

var HttpWebKey, NetType string

// LoadConfig read config from file and init dspFrontServer run environment variable
//  call before Start pool.Server()
// 	if return is not nil, should not start pool.Server()
func LoadConfig(confFile string) error {
	config, err := toml.LoadFile(confFile)
	if nil != err {
		log.Errorf("open config file [%v] failed[%v]!", confFile, err)
		return err
	}

	if err = initRedis(config); nil != err {
		log.Errorf("get Redis config failed[%v]!", err)
		return err
	}

	if err = initDB(config); nil != err {
		log.Errorf("get db config failed:[%v]!", err)
		return err
	}

	if err = initToken(config); nil != err {
		log.Errorf("get token config failed:[%v]!", err)
		return err
	}
	if err = initCommon(config); nil != err {
		log.Errorf("get common config failed:[%v]!", err)
		return err
	}

	return nil
}

// initRedis 初始化Redis连接
func initRedis(config *toml.Tree) error {

	redisInfo := struct {
		Addr     string
		Password string
		Db       int
		PoolSize int
	}{}
	redisInfo.Addr = config.GetDefault("Redis.host", "127.0.0.1:6379").(string)
	redisInfo.Password = config.GetDefault("Redis.pass", "127.0.0.1:6379").(string)
	redisInfo.Db = int(util.ToInt64(config.GetDefault("Redis.index", 0)))
	redisInfo.PoolSize = int(util.ToInt64(config.GetDefault("Redis.poolSize", 10)))

	RedisCli = redis.NewClient(redisInfo.Addr, redisInfo.Password, redisInfo.Db, redisInfo.PoolSize)

	return nil
}

//initDB 初始化DB baseAdapter.loadAdxTemplateData use
func initDB(config *toml.Tree) error {
	const NodeName = "mysql"
	host := config.GetDefault(fmt.Sprintf("%v.host", NodeName), "127.0.0.1").(string)
	port := config.GetDefault(fmt.Sprintf("%v.port", NodeName), "3306").(string)
	schema := config.GetDefault(fmt.Sprintf("%v.schema", NodeName), "tron").(string)
	user := config.GetDefault(fmt.Sprintf("%v.user", NodeName), "tron").(string)
	passwd := config.GetDefault(fmt.Sprintf("%v.pass", NodeName), "tron").(string)

	log.Debugf("host:%v, port:%v, schema:%v, user:%v, passwd:%v", host, port, schema, user, passwd)

	mysql.Initialize(host, port, schema, user, passwd)

	return nil
}

//initToken 初始化token参数
func initToken(config *toml.Tree) error {
	const NodeName = "token"
	DefaultPath = config.GetDefault(fmt.Sprintf("%v.defaultPath", NodeName), "/data/images/tokenLogo").(string)
	TokenTemplate = config.GetDefault(fmt.Sprintf("%v.tokenTemplate", NodeName), "/data/images/tokenTemplate/").(string)
	ImgURL = config.GetDefault(fmt.Sprintf("%v.imgURL", NodeName), "http://coin.top/tokenLogo").(string)
	TokenTemplateFile = config.GetDefault(fmt.Sprintf("%v.tokenTemplateFile", NodeName), "http://coin.top/tokenTemplate/TronscanTokenInformationSubmissionTemplate.xlsx").(string)
	log.Printf("defaultPath:[%v], tokenTemplate:[%v],imgURL:[%v],tokenTemplateFile:[%v]", DefaultPath, TokenTemplate, ImgURL, TokenTemplateFile)

	return nil
}

//initCommon 初始化common参数
func initCommon(config *toml.Tree) error {
	const NodeName = "common"
	HttpWebKey = config.GetDefault(fmt.Sprintf("%v.httpWebKey", NodeName), "WoiYeI5brZy4S8wQfVz7M5BczMkIhnugYW5QIibNgnWsAsktgHn5").(string)
	NetType = config.GetDefault(fmt.Sprintf("%v.netType", NodeName), "testnet").(string)
	log.Printf("HttpWebKey:[%v], NetType:[%v]", HttpWebKey, NetType)

	return nil
}
