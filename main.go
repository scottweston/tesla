package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jsgoecke/tesla"
	"github.com/spf13/viper"
	"gopkg.in/redis.v4"
)

func makeTimeStamp() string {
	return strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
}

func storeStatusInRedis(redis_client *redis.Client, chargeState *tesla.ChargeState) error {
	if chargeState == nil || redis_client == nil {
		return errors.New("nil argument passed")
	}
	stringChargeState := make(map[string]string)
	stringChargeState["When"] = makeTimeStamp()
	stringChargeState["ChargingState"] = chargeState.ChargingState
	stringChargeState["ChargeLimitSoc"] = strconv.Itoa(chargeState.ChargeLimitSoc)
	stringChargeState["BatteryLevel"] = strconv.Itoa(chargeState.BatteryLevel)
	stringChargeState["UsableBatteryLevel"] = strconv.Itoa(chargeState.UsableBatteryLevel)
	stringChargeState["ChargeCurrentRequest"] = strconv.Itoa(chargeState.ChargeCurrentRequest)
	stringChargeState["ChargeCurrentRequestMax"] = strconv.Itoa(chargeState.ChargeCurrentRequestMax)
	stringChargeState["BatteryRange"] = strconv.FormatFloat(chargeState.BatteryRange*1.60934, 'f', 2, 64)
	stringChargeState["EstBatteryRange"] = strconv.FormatFloat(chargeState.EstBatteryRange*1.60934, 'f', 2, 64)
	stringChargeState["IdealBatteryRange"] = strconv.FormatFloat(chargeState.IdealBatteryRange*1.60934, 'f', 2, 64)
	stringChargeState["ChargeEnergyAdded"] = strconv.FormatFloat(chargeState.ChargeEnergyAdded, 'f', 2, 64)
	stringChargeState["ChargeRangeAddedRated"] = strconv.FormatFloat(chargeState.ChargeMilesAddedRated*1.60934, 'f', 2, 64)
	stringChargeState["ChargeRangeAddedIdeal"] = strconv.FormatFloat(chargeState.ChargeMilesAddedIdeal*1.60934, 'f', 2, 64)
	stringChargeState["TimeToFullCharge"] = strconv.FormatFloat(chargeState.TimeToFullCharge, 'f', 2, 64)
	stringChargeState["ChargeRate"] = strconv.FormatFloat(chargeState.ChargeRate, 'f', 2, 64)
	stringChargeState["ChargePortDoorOpen"] = strconv.FormatBool(chargeState.ChargePortDoorOpen)
	stringChargeState["ChargePortLatch"] = chargeState.ChargePortLatch

	_, err := redis_client.HMSet("tesla_state", stringChargeState).Result()
	if err != nil {
		return (err)
	}
	return (nil)
}

func call_tesla(conf *viper.Viper, redis_client *redis.Client) error {
	client, err := tesla.NewClient(
		&tesla.Auth{
			ClientID:     conf.GetString("client_id"),
			ClientSecret: conf.GetString("client_secret"),
			Email:        conf.GetString("username"),
			Password:     conf.GetString("password"),
		})
	if err != nil {
		return err
	}

	vehicles, err := client.Vehicles()
	if err != nil {
		return err
	}
	vehicle := vehicles[conf.GetInt("vehicle")]

	if len(os.Args) == 2 {
		if os.Args[1] == "version" {
			log.Println("version 1.1")
			return nil
		} else if os.Args[1] == "start" {
			err := vehicle.StartCharging()
			return err
		} else if os.Args[1] == "stop" {
			err := vehicle.StopCharging()
			return err
		} else if os.Args[1] == "honk" {
			err := vehicle.HonkHorn()
			err = vehicle.HonkHorn()
			return err
		} else if os.Args[1] == "lock" {
			err := vehicle.LockDoors()
			return err
		} else if os.Args[1] == "boot" {
			err := vehicle.OpenTrunk("rear")
			return err
		} else if os.Args[1] == "unlock" {
			err := vehicle.UnlockDoors()
			return err
		} else if os.Args[1] == "state" || os.Args[1] == "status" {
			chargeState, err := vehicle.ChargeState()
			if err != nil {
				return err
			}
			data, err := json.Marshal(chargeState)
			if err != nil {
				return err
			}
			if data != nil {
				fmt.Println(string(data))
			}
			if redis_client != nil {
				err = storeStatusInRedis(redis_client, chargeState)
				if err != nil {
					return err
				}
			}
			return nil
		} else {
			i, err := strconv.Atoi(os.Args[1])
			if err != nil {
				log.Fatal("Usage: tesla [state|start|stop|50-100]")
				return nil
			} else {
				chargeState, err := vehicle.ChargeState()
				if err != nil {
					return err
				}
				if chargeState.ChargeLimitSoc != i {
					err := vehicle.SetChargeLimit(i)
					if err != nil {
						return err
					}
					chargeState, err = vehicle.ChargeState()
					if err != nil {
						return err
					}
				}
				data, err := json.Marshal(chargeState)
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				if redis_client != nil {
					err = storeStatusInRedis(redis_client, chargeState)
					if err != nil {
						return err
					}
				}
				return nil
			}
		}
	} else {
		log.Fatal("Usage: tesla [state|start|stop|50-100]")
		return nil
	}
}

func main() {
	conf := viper.New()

	conf.SetEnvPrefix("tesla")
	conf.SetConfigName("tesla")
	conf.AddConfigPath(".")
	conf.AddConfigPath("$HOME/.config/")

	conf.SetDefault("retries", 3)
	conf.SetDefault("vehicle", 0)
	conf.SetDefault("redis.host", "127.0.0.1")
	conf.SetDefault("redis.port", 6379)
	conf.SetDefault("redis.password", "")
	conf.SetDefault("redis.database", 0)

	err := conf.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
	redis_client := redis.NewClient(&redis.Options{
		Addr:     conf.GetString("redis.host") + ":" + conf.GetString("redis.port"),
		Password: conf.GetString("redis.password"),
		DB:       conf.GetInt("redis.database"),
	})

	_, err = redis_client.Ping().Result()
	if err != nil {
		redis_client = nil
		log.Printf("redis disabled: %s (%s:%d)", err, conf.GetString("redis.host"), conf.GetInt("redis.port"))
	}

	// The Tesla API endpoint can be unavailable quite often, this
	// will try and connect a few times
	var try = 0
	for try < conf.GetInt("retries") {
		err := call_tesla(conf, redis_client)
		if err == nil {
			break
		}
		log.Println(err)
		try++
	}
}
