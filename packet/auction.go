package packet

import (
	"fmt"

	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/econ"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/vsn"
)

type AuctionListItemsRequest struct {
	Auctioneer          guid.GUID
	ListFrom            uint32
	SearchedName        string
	LevelMin            uint8
	LevelMax            uint8
	AuctionSlotID       uint32
	AuctionMainCategory uint32
	AuctionSubCategory  uint32
	Quality             uint32
	Usable              uint8
}

type AuctionListing struct {
	ID                   uint32
	Entry                uint32
	EnchantmentID        uint32
	ItemRandomPropertyID uint32
	ItemSuffixFactor     uint32
	Owner                guid.GUID
	StartBid             econ.Money
	OutBid               econ.Money
	ExpireTime           uint32
	CurrentBidder        guid.GUID
	CurrentBid           econ.Money
}

func (al AuctionListing) Encode(version vsn.Build, e *etc.Buffer) {
	e.WriteUint32(al.ID)
	e.WriteUint32(al.Entry)
	e.WriteUint32(al.EnchantmentID)
	e.WriteUint32(al.ItemRandomPropertyID)
	e.WriteUint32(al.ItemSuffixFactor)
	al.Owner.EncodeUnpacked(version, e)
	e.WriteInt32(int32(al.StartBid))
	e.WriteInt32(int32(al.OutBid))
	e.WriteUint32(al.ExpireTime)
	al.CurrentBidder.EncodeUnpacked(version, e)
	e.WriteInt32(al.CurrentBid.Int32())
}

func DecodeAuctionListing(version vsn.Build, e *etc.Buffer) (AuctionListing, error) {
	al := AuctionListing{}
	al.ID = e.ReadUint32()
	al.Entry = e.ReadUint32()
	al.EnchantmentID = e.ReadUint32()
	al.ItemRandomPropertyID = e.ReadUint32()
	al.ItemSuffixFactor = e.ReadUint32()
	var err error
	al.Owner, err = guid.DecodeUnpacked(version, e)
	if err != nil {
		return al, err
	}
	al.StartBid = econ.Money(e.ReadInt32())
	al.OutBid = econ.Money(e.ReadInt32())
	al.ExpireTime = e.ReadUint32()
	al.CurrentBidder, err = guid.DecodeUnpacked(version, e)
	if err != nil {
		return al, err
	}

	al.CurrentBid = econ.Money(e.ReadInt32())
	return al, nil
}

type AuctionListItemsResult struct {
	Listings   []AuctionListing
	TotalCount uint32
}

func UnmarshalAuctionListItemsResult(version vsn.Build, b []byte) (*AuctionListItemsResult, error) {
	const maxPage = 50

	alir := new(AuctionListItemsResult)
	e := etc.FromBytes(b)

	count := e.ReadUint32()

	if count > maxPage {
		return nil, fmt.Errorf("server sent more auction house listings than expected")
	}

	alir.Listings = make([]AuctionListing, count)

	var err error
	var listing AuctionListing

	for l := uint32(0); l < count; l++ {
		listing, err = DecodeAuctionListing(version, e)
		if err != nil {
			return nil, err
		}

		alir.Listings[l] = listing
	}

	alir.TotalCount = e.ReadUint32()
	return alir, nil
}
