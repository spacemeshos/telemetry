package toolkit

import (
	"github.com/intel-go/cpuid"
	"github.com/spacemeshos/telemetry"
	"runtime"
	"strings"
)

func ConstantCpuInfo(prefix string, channels ...*telemetry.ChannelObject) {
	ifthen := func(a bool, b float64, c float64) float64 {
		if a {
			return b
		}
		return c
	}
	telemetry.ConstantString(prefix+".vendor", cpuid.VendorIdentificatorString, channels...)
	telemetry.ConstantString(prefix+".brand", strings.TrimRight(cpuid.ProcessorBrandString, "\u0000"), channels...)
	htt := cpuid.HasFeature(cpuid.HTT)
	telemetry.ConstantNumber(prefix+".cores", float64(runtime.NumCPU())/ifthen(htt, 2, 1), channels...)
	telemetry.ConstantNumber(prefix+".threads", float64(runtime.NumCPU()), channels...)
	telemetry.ConstantNumber(prefix+".AVX", ifthen(cpuid.HasFeature(cpuid.AVX), 1, 0), channels...)
	telemetry.ConstantNumber(prefix+".AVX2", ifthen(cpuid.HasExtendedFeature(cpuid.AVX2), 1, 0), channels...)
	telemetry.ConstantNumber(prefix+".AES", ifthen(cpuid.HasFeature(cpuid.AES), 1, 0), channels...)
}
