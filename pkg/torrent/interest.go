package torrent

func (c *Connection) setInterested(interested bool) error {
	if c.localInterested == interested {
		return nil
	}

	var id uint8 = PacketNotInterested
	if interested {
		id = PacketInterested
	}

	err := c.WritePacket(&Packet{
		ID: id,
	})
	if err != nil {
		return err
	}
	c.localInterested = interested
	return nil
}
