package generator

type UserConfig struct {
	Name string `yaml:"Name"`
}

type CAConfig struct {
	CommonName string `yaml:"CommonName"`
	Country    string `yaml:"Country"`
	Province   string `yaml:"Province"`
	Locality   string `yaml:"Locality"`
	Expire     int    `yaml:"Expire"`
}

type CertConfig struct {
	Name string       `yaml:"Name"`
	CA   CAConfig     `yaml:"CA"`
	User []UserConfig `yaml:"User"`
}

type GenConfig struct {
	SignType      string       `yaml:"SignType"`
	Root          CertConfig   `yaml:"Root"`
	Organizations []CertConfig `yaml:"Organizations"`
}

func (cfg *GenConfig) GetOrgCertConfig(orgName string) *CertConfig {
	for _, certCfg := range cfg.Organizations {
		if certCfg.Name == orgName {
			return &certCfg
		}
	}

	return nil
}
