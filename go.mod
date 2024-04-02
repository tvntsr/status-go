module github.com/status-im/status-go

go 1.21

replace github.com/prometheus/procfs v0.7.3 => ./replacements/prometheus/procfs
replace github.com/prometheus/client_golang v1.12.0  => ./replacements/prometheus/client_golang
replace github.com/prometheus/client_model v0.2.1-0.20210607210712-147c58e9608a => ./replacements/prometheus/client_model
replace github.com/prometheus/common v0.32.1 => ./replacements/prometheus/common
replace github.com/VictoriaMetrics/fastcache v1.12.1 => ./replacements/fastcache


require (
	github.com/Masterminds/squirrel v1.5.4
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/ethereum/go-ethereum v1.13.14
	github.com/golang/protobuf v1.5.4
	github.com/google/uuid v1.6.0
	github.com/jellydator/ttlcache/v3 v3.2.0
	github.com/jmoiron/sqlx v1.3.5
	github.com/ladydascalie/currency v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/status-im/status-go/extkeys v1.1.2
	github.com/stretchr/testify v1.9.0
	github.com/wk8/go-ordered-map/v2 v2.1.8
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/exp v0.0.0-20240325151524-a685a6edb6d8
	golang.org/x/net v0.22.0
	golang.org/x/text v0.14.0
)

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/VictoriaMetrics/fastcache v1.12.1 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.10.0 // indirect
	github.com/btcsuite/btcd v0.20.1-beta // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cockroachdb/errors v1.8.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20190617123548-eb05cc24525f // indirect
	github.com/cockroachdb/pebble v0.0.0-20230928194634-aa077af62593 // indirect
	github.com/cockroachdb/redact v1.0.8 // indirect
	github.com/cockroachdb/sentry-go v0.6.1-cockroachdb.2 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/consensys/bavard v0.1.13 // indirect
	github.com/consensys/gnark-crypto v0.12.1 // indirect
	github.com/crate-crypto/go-ipa v0.0.0-20231025140028-3c0104f4b233 // indirect
	github.com/crate-crypto/go-kzg-4844 v0.7.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deckarep/golang-set/v2 v2.1.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/ethereum/c-kzg-4844 v0.4.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gballet/go-verkle v0.1.1-0.20231031103413-a67434b50f46 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/holiman/bloomfilter/v2 v2.0.3 // indirect
	github.com/holiman/uint256 v1.2.4 // indirect
	github.com/klauspost/compress v1.15.15 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mmcloughlin/addchain v0.4.0 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.12.0 // indirect
	github.com/prometheus/client_model v0.2.1-0.20210607210712-147c58e9608a // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/shirou/gopsutil v3.21.4-0.20210419000835-c7a38de76ee5+incompatible // indirect
	github.com/supranational/blst v0.3.11 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/mod v0.16.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/tools v0.19.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	rsc.io/tmplfunc v0.0.3 // indirect
)
