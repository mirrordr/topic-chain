package utils

import (
	"context"
	"github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

var myclient *qdrant.Client
var targetCollectionName string
var threshold = 0.3
var qdrantHost = "127.0.0.1"
var port = 6334

func InitQDrantClient() {
	if myclient == nil {
		client, err := qdrant.NewClient(&qdrant.Config{
			Host: qdrantHost,
			Port: port,
			GrpcOptions: []grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100*1024*1024), grpc.MaxCallSendMsgSize(100*1024*1024)),
			},
		})
		if err != nil {
			panic("qdrant client unable to start" + err.Error())
		}
		myclient = client
		targetCollectionName = "temp-collection-" + time.Now().Format("20060102150405")
		err = myclient.CreateCollection(context.Background(), &qdrant.CreateCollection{
			CollectionName: targetCollectionName,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     1536,
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			panic("qdrant client unable to create collection: " + err.Error())
		}
	}
	log.Println("qdrant client start successfully, collection name is: ", targetCollectionName)

}

func TopicSimilaritySearchFromQDrant(vector []float32) (int, string) {
	searchResult, err := myclient.Query(context.Background(), &qdrant.QueryPoints{
		CollectionName: targetCollectionName,
		Query:          qdrant.NewQuery(vector...),
		WithPayload:    qdrant.NewWithPayloadInclude("topic"),
	})
	if err != nil {
		panic("qdrant error when upsert vector: " + err.Error())
	}

	for i := 0; i < len(searchResult); i++ {
		res := searchResult[i]
		log.Println(res.Score)
		log.Println(res.Id.GetNum())
		log.Println(res.Payload["topic"].String())
	}
	if searchResult[0].Score <= float32(threshold) {
		log.Println("【弃用】当前分数不满足阈值")
		return -1, ""
	}

	return int(searchResult[0].Id.GetNum()), searchResult[0].Payload["topic"].String()
}

func UpsertVectorToQDrant(id int, vector []float32, topic string) {
	req := &qdrant.UpsertPoints{}
	if id < 0 {
		req = &qdrant.UpsertPoints{
			CollectionName: targetCollectionName,
			Points: []*qdrant.PointStruct{
				{
					Vectors: qdrant.NewVectorsDense(vector),
					Payload: qdrant.NewValueMap(map[string]any{"topic": topic}),
				},
			},
		}
	} else {
		req = &qdrant.UpsertPoints{
			CollectionName: targetCollectionName,
			Points: []*qdrant.PointStruct{
				{
					Id:      qdrant.NewIDNum(uint64(id)),
					Vectors: qdrant.NewVectorsDense(vector),
					Payload: qdrant.NewValueMap(map[string]any{"topic": topic}),
				},
			},
		}
	}
	operationInfo, err := myclient.Upsert(context.Background(), req)
	if err != nil {
		panic("qdrant error when upsert vector: " + err.Error())
	}
	log.Println("upsert success: ", operationInfo.String())
}
