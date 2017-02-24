package consulmemcached

import (
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/netflix/rend/common"
)

func gomemcacheErrorMapper(err error) error {
	switch err {
	case memcache.ErrNoServers:
		return common.ErrInternal
	}
	return err
}
