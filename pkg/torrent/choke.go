package torrent

func (c *Connection) SetChoked(choked bool) error {
	if c.localChoking == choked {
		return nil
	}

	var id uint8 = PacketUnchoke
	if choked {
		id = PacketChoke
	}

	err := c.WritePacket(&Packet{
		ID: id,
	})
	if err != nil {
		return err
	}
	c.localChoking = choked
	return nil
}
