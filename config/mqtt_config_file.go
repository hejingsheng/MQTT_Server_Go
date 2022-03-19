package config

import "github.com/Unknwon/goconfig"

var cfg *goconfig.ConfigFile

func ConfigFileInit(file string) {
	config, err := goconfig.LoadConfigFile(file)
	if err != nil{
		panic("错误")
	}
	cfg = config
}

func ReadConfigValueInt(section string, key string) (int,error)  {
	value, err := cfg.Int(section, key)
	if err != nil {
		return 0,err
	}
	return value,nil
}

func ReadConfigValueString(section string, key string) (string,error)  {
	value, err := cfg.GetValue(section, key)
	if err != nil {
		return "",err
	}
	return value,nil
}

func ReadConfigValueBool(section string, key string) (bool,error)  {
	value, err := cfg.Bool(section, key)
	if err != nil {
		return false, err
	}
	return value,nil
}

func ReadConfigValueDouble(section string, key string) (float64,error)  {
	value, err := cfg.Float64(section, key)
	if err != nil {
		return 0.0, err
	}
	return value,nil
}
