package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	sunspec "github.com/andig/gosunspec"
	bus "github.com/andig/gosunspec/modbus"
	"github.com/andig/gosunspec/smdx"
	"github.com/volkszaehler/mbmd/meters"
	quirks "github.com/volkszaehler/mbmd/meters/sunspec"
)

func modelName(m sunspec.Model) string {
	model := smdx.GetModel(uint16(m.Id()))
	if model == nil {
		return ""
	}
	return model.Name
}

func main() {

	conn := meters.NewTCP("192.168.69.10:1502")
	conn.Timeout(15 * time.Second)
	conn.Slave(1) // modbus device id
	defer conn.Close()

	in, err := bus.Open(conn.ModbusClient())
	if err != nil && in == nil {
		log.Fatalln(err)
	} else if err != nil {
		log.Printf("warning: device opened with partial result: %v", err) // log error but continue
	}

	var Mn, Md, SN string
	in.Do(func(d sunspec.Device) {
		d.Do(func(m sunspec.Model) {
			// fmt.Println("Model:", m.Id(), modelName(m))
			blocknum := 0
			m.Do(func(b sunspec.Block) {
				if blocknum > 0 {
					fmt.Println("Block:", blocknum)
				}
				blocknum++

				err = b.Read()
				if err != nil {
					log.Printf("skipping due to read error: %v", err)
					return
				}

				b.Do(func(p sunspec.Point) {
					t := p.Type()[0:3]
					v := p.Value()
					if p.NotImplemented() {
						v = "n/a"
					} else if t == "int" || t == "uin" || t == "acc" {
						quirks.FixKostal(p)
						v = p.ScaledValue()
						v = fmt.Sprintf("%.2f", v)
					}

					vs := fmt.Sprintf("%17v", v)

					if m.Id() == 1 {
						switch p.Id() {
						case "Mn":
							Mn = vs
						case "Md":
							Md = vs
						case "SN":
							SN = vs
						}
					} else {

						// fmt.Printf("%s\t%s\t   %s\n", p.Id(), vs, p.Type())
						// if p.Id() == "W" {
						fmt.Printf("%s %s model: \"%s\" (SN: %s)\t%s = %s\n", strings.TrimSpace(Mn), modelName(m), strings.TrimSpace(Md), strings.TrimSpace(SN), p.Id(), vs)
						// }
					}

				})
			})
		})
	})

}
