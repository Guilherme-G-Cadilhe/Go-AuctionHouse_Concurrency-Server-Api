package bid

import (
	"context"
	"fmt"
	"time"

	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/configuration/logger"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/entity/bid_entity"
	"github.com/Guilherme-G-Cadilhe/Go-AuctionHouse_Concurrency-Server-Api/internal/internal_error"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (bd *BidRepository) FindBidByAuctionId(ctx context.Context, auctionId string) ([]bid_entity.Bid, *internal_error.InternalError) {
	filter := bson.M{"auction_id": auctionId}

	var bids []BidEntityMongo
	cursor, err := bd.Collection.Find(ctx, filter)
	if err != nil {
		logger.Error(fmt.Sprintf("error trying to find bids by auction id %s", auctionId), err)
		return nil, internal_error.NewInternalServerError(fmt.Sprintf("error trying to find bids by auction id %s", auctionId))
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &bids); err != nil {
		logger.Error(fmt.Sprintf("error trying to find bids by auction id %s", auctionId), err)
		return nil, internal_error.NewInternalServerError(fmt.Sprintf("error trying to find bids by auction id %s", auctionId))
	}
	bidsEntities := make([]bid_entity.Bid, len(bids))
	for i, bid := range bids {
		bidsEntities[i] = bid_entity.Bid{
			Id:        bid.Id,
			UserId:    bid.UserId,
			AuctionId: bid.AuctionId,
			Amount:    bid.Amount,
			Timestamp: time.Unix(bid.Timestamp, 0),
		}
	}
	return bidsEntities, nil
}

func (bd *BidRepository) FindWinningBidByAuctionId(ctx context.Context, auctionId string) (*bid_entity.Bid, *internal_error.InternalError) {
	filter := bson.M{"auction_id": auctionId}

	opts := options.FindOne().SetSort(bson.D{{Key: "amount", Value: -1}})

	var bid BidEntityMongo
	err := bd.Collection.FindOne(ctx, filter, opts).Decode(&bid)
	if err != nil {
		logger.Error(fmt.Sprintf("error trying to find winning bid by auction id %s", auctionId), err)
		return nil, internal_error.NewNotFoundError(fmt.Sprintf("error trying to find winning bid by auction id %s", auctionId))
	}
	return &bid_entity.Bid{
		Id:        bid.Id,
		UserId:    bid.UserId,
		AuctionId: bid.AuctionId,
		Amount:    bid.Amount,
		Timestamp: time.Unix(bid.Timestamp, 0),
	}, nil
}
