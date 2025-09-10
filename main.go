package main

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"

	"encbench/pb"

	"github.com/andybalholm/brotli"
	"google.golang.org/protobuf/proto"
)

type JRecord struct {
	ID   int     `json:"id"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	V1   float64 `json:"v1"`
	V2   float64 `json:"v2"`
	V3   float64 `json:"v3"`
	Name string  `json:"name"`
}

func makeData() ([]JRecord, *pb.Data) {
	j := make([]JRecord, 0, 12)
	p := &pb.Data{Records: make([]*pb.Record, 0, 12)}
	for i := 0; i < 12; i++ {
		rj := JRecord{
			ID:   i + 1,
			X:    float64(i) * 1.25,
			Y:    float64(i)*1.25 + 0.5,
			V1:   3.14159 * float64(i+1),
			V2:   2.71828 * float64(12-i),
			V3:   float64((i+1)*(i+2)) / 7.0,
			Name: fmt.Sprintf("item-%02d", i+1),
		}
		j = append(j, rj)
		p.Records = append(p.Records, &pb.Record{
			Id:   int32(rj.ID),
			X:    rj.X,
			Y:    rj.Y,
			V1:   rj.V1,
			V2:   rj.V2,
			V3:   rj.V3,
			Name: rj.Name,
		})
	}
	return j, p
}

type sizes struct {
	raw     int
	gzip    int
	deflate int
	br      int
}

func compressAll(label string, raw []byte) sizes {
	var out sizes
	out.raw = len(raw)

	// gzip (BestCompression)
	{
		var buf bytes.Buffer
		zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := zw.Write(raw); err != nil {
			log.Fatal(err)
		}
		if err := zw.Close(); err != nil {
			log.Fatal(err)
		}
		out.gzip = buf.Len()
	}

	// raw DEFLATE (RFC1951) via compress/flate (no zlib header)
	{
		var buf bytes.Buffer
		zw, err := flate.NewWriter(&buf, flate.BestCompression)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := zw.Write(raw); err != nil {
			log.Fatal(err)
		}
		if err := zw.Close(); err != nil {
			log.Fatal(err)
		}
		out.deflate = buf.Len()
	}

	// Brotli (level 11)
	{
		var buf bytes.Buffer
		zw := brotli.NewWriterLevel(&buf, 11)
		if _, err := zw.Write(raw); err != nil {
			log.Fatal(err)
		}
		if err := zw.Close(); err != nil {
			log.Fatal(err)
		}
		out.br = buf.Len()
	}

	fmt.Printf("%-10s  raw=%4d  gzip=%4d  deflate=%4d  brotli=%4d\n",
		label, out.raw, out.gzip, out.deflate, out.br)
	return out
}

func main() {
	jd, pd := makeData()

	// JSON encode
	jsonBytes, err := json.MarshalIndent(jd, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	// Protobuf encode
	protoBytes, err := proto.Marshal(pd)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Byte sizes (smaller is better):")
	_ = compressAll("json", jsonBytes)
	_ = compressAll("protobuf", protoBytes)
}
