module huobicagent

go 1.17

replace tczbgo => git.zbdigital.net/architecture/tczbgo.git v0.0.44

replace cryptopackage => git.zbdigital.net/currency/cryptopackage.git v0.0.7

require (
	github.com/PuerkitoBio/goquery v1.8.0
	github.com/go-rod/rod v0.101.8
)

require (
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/segmentio/kafka-go v0.4.25 // indirect
)

require (
	cryptopackage v0.0.7
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/ysmood/goob v0.3.0 // indirect
	github.com/ysmood/gson v0.6.4 // indirect
	github.com/ysmood/leakless v0.7.0 // indirect
	golang.org/x/net v0.0.0-20210916014120-12bc252f5db8 // indirect
	tczbgo v0.0.44
)
