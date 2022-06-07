package georedis

import (
	"testing"

	geo2 "github.com/henrion-y/api-base/infra/geo"

	"github.com/spf13/viper"
)

func getCache() geo2.Geo {
	conf := viper.New()

	cacheRdb, err := NewRedisProvider(conf)
	if err != nil {
		panic(err)
	}
	return cacheRdb
}

func TestGeo_Add(t *testing.T) {
	rdb := getCache()
	data := []*geo2.Member{
		{
			Name: "shenzhen",
			Coordinate: geo2.Coordinate{
				Lat: 113.88308,
				Lon: 22.55329,
			},
		},
		{
			Name: "dongguan",
			Coordinate: geo2.Coordinate{
				Lat: 113.75,
				Lon: 23.05,
			},
		},
		{
			Name: "changsha",
			Coordinate: geo2.Coordinate{
				Lat: 112.98626,
				Lon: 28.25591,
			},
		},
		{
			Name: "fuyang",
			Coordinate: geo2.Coordinate{
				Lat: 115.80,
				Lon: 32.91,
			},
		},
	}
	err := rdb.Add("geo:city", data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGeo_Pos(t *testing.T) {
	rdb := getCache()
	members, err := rdb.Pos("geo:city", "fuyang", "123", "changsha")
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < len(members); i++ {
		t.Log(members[i])
	}
}

func TestGeo_Dist(t *testing.T) {
	rdb := getCache()

	dist, err := rdb.Dist("geo:city", "shenzhen", "dongguan", geo2.KM)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(dist)

	dist, err = rdb.Dist("geo:city", "shenzhen", "123", geo2.KM)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(dist)
}

func TestGeo_RadiusByName(t *testing.T) {
	rdb := getCache()

	members, err := rdb.RadiusByName("geo:city", "shenzhen", 6000, geo2.KM, geo2.WithDist)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < len(members); i++ {
		t.Log(members[i])
	}
}

func TestGeo_Del(t *testing.T) {
	rdb := getCache()
	err := rdb.Del("geo:city", "fuyang")
	if err != nil {
		t.Fatal(err)
	}
}
