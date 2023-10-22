package announce

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/Doridian/foxTorrent/pkg/bencoding"
)

func Decode(data []byte) (*Announce, error) {
	decoded, err := bencoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	announce := &Announce{}

	decodedDict, ok := decoded.(map[string]interface{})
	if !ok {
		return nil, bencoding.ErrInvalidType
	}

	failureReasonRaw, ok := decodedDict["failure reason"]
	if ok { // optional
		failureReasonTyped, ok := failureReasonRaw.([]byte)
		if !ok {
			return nil, bencoding.ErrInvalidType
		}
		return nil, fmt.Errorf("tracker failure: %s", failureReasonTyped)
	}

	warningMessageRaw, ok := decodedDict["warning message"]
	if ok { // optional
		warningMessageTyped, ok := warningMessageRaw.([]byte)
		if !ok {
			return nil, bencoding.ErrInvalidType
		}
		announce.WarningMessage = string(warningMessageTyped)
	}

	trackerIDRaw, ok := decodedDict["tracker id"]
	if ok { // optional
		trackerIDTyped, ok := trackerIDRaw.([]byte)
		if !ok {
			return nil, bencoding.ErrInvalidType
		}
		announce.TrackerID = string(trackerIDTyped)
	}

	intervalRaw, ok := decodedDict["interval"]
	if !ok { // required
		return nil, bencoding.ErrMissingRequiredField
	}
	intervalTyped, ok := intervalRaw.(int64)
	if !ok {
		return nil, bencoding.ErrInvalidType
	}
	announce.Interval = uint32(intervalTyped)

	minIntervalRaw, ok := decodedDict["min interval"]
	if ok { // optional
		minIntervalTyped, ok := minIntervalRaw.(int64)
		if !ok {
			return nil, bencoding.ErrInvalidType
		}
		announce.MinInterval = uint32(minIntervalTyped)
	}

	completeRaw, ok := decodedDict["complete"]
	if !ok { // required
		return nil, bencoding.ErrMissingRequiredField
	}
	completeTyped, ok := completeRaw.(int64)
	if !ok {
		return nil, bencoding.ErrInvalidType
	}
	announce.Complete = uint32(completeTyped)

	incompleteRaw, ok := decodedDict["incomplete"]
	if !ok { // required
		return nil, bencoding.ErrMissingRequiredField
	}
	incompleteTyped, ok := incompleteRaw.(int64)
	if !ok {
		return nil, bencoding.ErrInvalidType
	}
	announce.Incomplete = uint32(incompleteTyped)

	peersRaw, ok := decodedDict["peers"]
	if !ok { // required
		return nil, bencoding.ErrMissingRequiredField
	}

	peersTyped, ok := peersRaw.([]interface{})
	if ok { // dict model
		peers := make([]Peer, 0, len(peersTyped))
		for _, peerRaw := range peersTyped {
			peer := Peer{}
			peerDict, ok := peerRaw.(map[string]interface{})
			if !ok {
				return nil, bencoding.ErrInvalidType
			}

			peerIdRaw, ok := peerDict["peer id"]
			if ok { // optional
				peerIdTyped, ok := peerIdRaw.([]byte)
				if !ok {
					return nil, bencoding.ErrInvalidType
				}
				peer.PeerID = string(peerIdTyped)
			}

			ipRaw, ok := peerDict["ip"]
			if !ok { // required
				return nil, bencoding.ErrMissingRequiredField
			}
			ipTyped, ok := ipRaw.([]byte)
			if !ok {
				return nil, bencoding.ErrInvalidType
			}
			peer.IP = net.ParseIP(string(ipTyped))
			if peer.IP == nil {
				return nil, fmt.Errorf("invalid IP address: %s", ipTyped)
			}

			portRaw, ok := peerDict["port"]
			if !ok { // required
				return nil, bencoding.ErrMissingRequiredField
			}
			portTyped, ok := portRaw.(int64)
			if !ok {
				return nil, bencoding.ErrInvalidType
			}
			peer.Port = uint16(portTyped)

			peers = append(peers, peer)
		}
		announce.Peers = peers
	} else { // binary model
		peersTypedBinary, ok := peersRaw.([]byte)
		if !ok {
			return nil, bencoding.ErrInvalidType
		}

		if len(peersTypedBinary)%6 != 0 {
			return nil, fmt.Errorf("invalid binary peers length: %d", len(peersTypedBinary))
		}

		peers := make([]Peer, 0, len(peersTypedBinary)/6)
		for i := 0; i < len(peersTypedBinary); i += 6 {
			peers = append(peers, Peer{
				IP:   net.IP(peersTypedBinary[i : i+4]),
				Port: binary.BigEndian.Uint16(peersTypedBinary[i+4 : i+6]),
			})
		}
		announce.Peers = peers
	}

	return announce, nil
}
