package util

import (
	"errors"
	"net"
	"sort"
	"strings"
	"time"
)

// CacheData represents a simple storage struct.
type CacheData struct {
	//MethodIdentifier identifies CacheData using on of the CRUD operations can be [GET, POST, UPDATE, DELETE]
	MethodIdentifier string
	//URL identifies which endpoint has been requested
	URL string
	//Protocol identifies which html protocol usually HTTP/1.1 [added it so that i'll reject other protocols]
	Protocol string
	//ResponseBody will be the data that will be cached/saved initially from the remote service.
	ResponseBody []byte
	//ID will be an identifier for the program to sync caching to and from client request to remote service.
	ID         int
	Expiration time.Time
}

// Cache represents the whole caching data been saved.
var Cache []CacheData

//CacheExpiration for expiration of caching in seconds
var CacheExpiration = 4 //4 seconds for the caching to expire

//TCPAddressResolver resolved an address and returns to an struct having ip and port.
func TCPAddressResolver(addr string) (tcpAddress *net.TCPAddr, err error) {
	tcpAddress, err = net.ResolveTCPAddr("tcp", addr)
	return
}

//AddCacheData will save data from remote service.
func (cacheData CacheData) AddCacheData() {
	Cache = append(Cache, cacheData)
}

//ExtractData extracts data from recieved data.
func ExtractData(data []byte, identifier int) (cacheData CacheData, err error) {
	dataStrings := strings.Split(string(data), "\n")
	baseInfo := strings.Split(dataStrings[0], " ")
	crud := strings.TrimSpace(baseInfo[0])
	//Cache only for GET method
	if crud != "GET" {
		err = errors.New("Caching only available for GET method")
		return
	}

	cacheData.MethodIdentifier = crud
	cacheData.URL = strings.TrimSpace(baseInfo[1])
	cacheData.Protocol = strings.TrimSpace(baseInfo[2])
	cacheData.ID = identifier
	cacheData.Expiration = time.Now()
	return
}

//ExtractDataFromRemote extracts data from recieved data.
func ExtractDataFromRemote(data []byte, identifier int) (cacheData CacheData, err error) {
	//This method first needs to check if there's the same URL based cache available in the Cache storage.
	dataStrings := strings.Split(string(data), "\n")
	baseInfo := strings.Split(dataStrings[0], " ")
	crud := strings.TrimSpace(baseInfo[0])
	//Cache only for GET method
	if crud != "GET" {
		err = errors.New("Caching only available for GET method")
		return
	}

	cacheData.MethodIdentifier = crud
	cacheData.URL = strings.TrimSpace(baseInfo[1])
	cacheData.Protocol = strings.TrimSpace(baseInfo[2])
	cacheData.ID = identifier
	return
}

// DoesCacheDataExistNB Tries to find a specific cacheData in Cache
func (cacheData *CacheData) DoesCacheDataExistNB() (idx int) {
	idx = -1
	if len(Cache) == 0 {
		return idx
	}

	if len(Cache) != 1 {
		sort.Slice(Cache, func(i, j int) bool { return Cache[i].URL < Cache[j].URL })

		idx = sort.Search(len(Cache), func(i int) bool {
			return string(Cache[i].URL) >= cacheData.URL
		})

		if idx == -1 {
			return
		}
	} else {
		idx = 0
	}

	if len(Cache[idx].ResponseBody) == 0 {
		idx = -1
		return
	}
	delay := time.Now().Sub(Cache[idx].Expiration).Seconds()
	if int(delay) > CacheExpiration {
		//Expired
		Cache[idx].ResponseBody = nil
		idx = -1
	}
	return
}

// SaveData Tries to find a specific cacheData in Cache
func (cacheData *CacheData) SaveData(data []byte) {
	sort.SliceStable(Cache, func(i, j int) bool { return Cache[i].ID < Cache[j].ID })
	idx := sort.Search(len(Cache), func(i int) bool {
		return Cache[i].ID >= cacheData.ID
	})
	if idx >= 0 && len(Cache[idx].ResponseBody) == 0 {
		Cache[idx].ResponseBody = data
		Cache[idx].Expiration = time.Now()
	}
	return
}

// GetCacheDataUsingURL Tries to find a specific cacheData in Cache
func GetCacheDataUsingURL(url string) (cacheData CacheData, err error) {
	if len(Cache) == 0 {
		err = errors.New("Empty caching")
		return
	}

	if len(Cache) == 1 {
		cacheData = Cache[0]
		return
	}
	sort.SliceStable(Cache, func(i, j int) bool { return Cache[i].URL < Cache[j].URL })
	idx := sort.Search(len(Cache), func(i int) bool {
		return Cache[i].URL >= url
	})
	if idx == -1 {
		err = errors.New("Not found")
		return
	}
	cacheData = Cache[idx]
	return
}

//GetCacheData returns CacheData using index
func GetCacheData(index int) (cacheData CacheData, err error) {
	if len(Cache) == 0 {
		err = errors.New("Empty Cache")
		return
	}
	cacheData = Cache[index]
	return
}
