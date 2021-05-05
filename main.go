package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

// 宣告
const (
	LED   = rpio.Pin(19) // GPIO#19 <=> PIN#35
	SERVO = rpio.Pin(12) // GPIO#12 <=> PIN#32
)

func main() {
	err := rpio.Open()
	if err != nil {
		log.Fatalln("Open GPIO Error: ", err)
	}
	defer func() {
		err := rpio.Close()
		if err != nil {
			log.Println("Close GPIO Error: ", err)
		}
	}()

	log.Println("Starting Pi Service...")

	LED.Output()    // 設置輸出腳位
	LED.Low()       // 先關閉LED
	defer LED.Low() // 結束程式時也關閉LED

	const ServoFreq = 50              // 伺服馬達頻率: 50Hz
	const cycleBase = 10000           // DutyCycle的基本長度
	var pwm uint32 = 900              // 預設PWM數值
	SERVO.Pwm()                       // 設置PWM腳位
	SERVO.Freq(ServoFreq * cycleBase) // 設置頻率
	time.Sleep(time.Second)
	SERVO.DutyCycle(pwm, cycleBase) // 先預設位置
	defer func() {
		SERVO.Output() // 結束程式時解除PWM模式
	}()

	/* 50Hz 每個角度對應的Duty值
	 *  176 / 10000 =  1.76% 最右邊 [v]
	 *  350 / 10000 =  3.50% +90˚  [v]
	 *  604 / 10000 =  6.04% +45˚  [v]
	 *  844 / 10000 =  8.44% 0˚    [v]
	 * 1114 / 10000 = 11.24% -45˚  [v]
	 * 1114 / 10000 = 11.24% -45˚  [v]
	 * 1324 / 10000 = 13.24% -90˚  [v]
	 */

	turnMotor := func(pwm uint32) {
		log.Println("現在 = ", pwm)
		SERVO.DutyCycle(pwm, cycleBase)
	}

	go func() {
		var cmd string
		for {
			cmd = ""
			fmt.Scanln(&cmd)
			cmd = strings.TrimSpace(cmd)
			log.Println("Scan => ", cmd)

			switch cmd {
			case "H", "h":
				LED.High()
			case "L", "l":
				LED.Low()
			case "A", "a":
				pwm += 50
				if pwm > 1500 {
					pwm = 1500
				}
				turnMotor(pwm)
			case "d", "D":
				pwm -= 50
				if pwm < 250 {
					pwm = 250
				}
				turnMotor(pwm)
			default:
				d, err := strconv.ParseUint(cmd, 10, 32)
				if err == nil {
					turnMotor(uint32(d))
				}
			}
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
