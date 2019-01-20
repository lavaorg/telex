package zfs

type Sysctl func(metric string) ([]string, error)
type Zpool func() ([]string, error)

type Zfs struct {
	KstatPath    string
	KstatMetrics []string
	PoolMetrics  bool
	sysctl       Sysctl
	zpool        Zpool
}
