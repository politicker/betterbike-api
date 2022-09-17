package citibike

func (b *Ebike) MaxDistance() float32 {
	return float32(b.BatteryStatus.DistanceRemaining.Value) / (float32(b.BatteryStatus.Percent) / 100)
}
