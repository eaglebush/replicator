package replicator

import (
	"log"
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
	cfg "github.com/eaglebush/config"
	"github.com/eaglebush/datahelper"
)

func TestInitReplicator(t *testing.T) {
	var r Replicator
	r = Replicator{}

	config, err := cfg.LoadConfig(`./config.json`)
	if err != nil {
		log.Fatal("Configuration file not found!")
	}

	// connect to database
	dh := datahelper.NewDataHelper(config)
	_, err = dh.Connect()
	defer dh.Disconnect(false)
	if err != nil {
		return
	}

	// define table
	fl := false
	err = r.Init(dh, `frt.freight`, []Column{
		{
			Name: `freight_key`,
			Type: `nvarchar(38)`,
			Null: &fl,
		},
		{
			Name: `ref_address`,
			Type: `nvarchar(400)`,
			Null: &fl,
		},
		{
			Name: `ref_amount`,
			Type: `decimal(18,3)`,
			Null: &fl,
		},
	}, []string{
		`freight_key`,
	})

	if err != nil {
		log.Printf("Error found: %s\r\n", err.Error())
	}
}

func TestInsertReplicator(t *testing.T) {
	var r Replicator
	r = Replicator{}

	config, err := cfg.LoadConfig(`./config.json`)
	if err != nil {
		log.Fatal("Configuration file not found!")
	}

	// connect to database
	dh := datahelper.NewDataHelper(config)
	_, err = dh.Connect()
	defer dh.Disconnect(false)
	if err != nil {
		return
	}

	// define table
	fl := false
	err = r.Init(dh, `frt.freight.init`, []Column{
		{
			Name: `freight_key`,
			Type: `nchar(38)`,
			Null: &fl,
		},
		{
			Name: `ref_address`,
			Type: `nvarchar(400)`,
			Null: &fl,
		},
		{
			Name: `ref_amount`,
			Type: `decimal(18,3)`,
			Null: &fl,
		},
		{
			Name: `priority`,
			Type: `int`,
			Null: &fl,
		},
		{
			Name: `priority2`,
			Type: `int`,
			Null: &fl,
		},
	}, []string{
		`freight_key`,
	})

	if err != nil {
		log.Printf("Error found: %s\r\n", err.Error())
	}

	data := `{
				"freight_key": "1QuIyNTdgyCQtxHeVuHxROrtw3B",
				"ref_address": "21 Jump St. Manila",
				"ref_amount": 1.10,
				"priority": 3,
				"priority2": -3
			 }`

	err = r.Insert(dh, `frt.freight.created`, []byte(data))
	if err != nil {
		log.Printf("Error found: %s\r\n", err.Error())
	}
}

func TestUpdateReplicator(t *testing.T) {
	var r Replicator
	r = Replicator{}

	config, err := cfg.LoadConfig(`./config.json`)
	if err != nil {
		log.Fatal("Configuration file not found!")
	}

	// connect to database
	dh := datahelper.NewDataHelper(config)
	_, err = dh.Connect()
	defer dh.Disconnect(false)
	if err != nil {
		return
	}

	// define table
	tr := true
	fl := false
	err = r.Init(dh, `frt.freight.init`, []Column{
		{
			Name: `freight_key`,
			Type: `nchar(38)`,
			Null: &fl,
		},
		{
			Name: `ref_address`,
			Type: `nvarchar(400)`,
			Null: &tr,
		},
		{
			Name: `ref_amount`,
			Type: `decimal(18,3)`,
			Null: &tr,
		},
	}, []string{
		`freight_key`,
	})

	if err != nil {
		log.Printf("Error found: %s\r\n", err.Error())
	}

	data := `{
				"freight_key": "1QuIyNTdgyCQtxHeVuHxROrtw3B",
				"ref_address": "21 Jump St. Quezon City",
				"ref_amount": 1.12
			 }`

	err = r.Update(dh, `frt.freight.updated`, []byte(data))
	if err != nil {
		log.Printf("Error found: %s\r\n", err.Error())
	}
}

func TestDeleteReplicator(t *testing.T) {
	var r Replicator
	r = Replicator{}

	config, err := cfg.LoadConfig(`./config.json`)
	if err != nil {
		log.Fatal("Configuration file not found!")
	}

	// connect to database
	dh := datahelper.NewDataHelper(config)
	_, err = dh.Connect()
	defer dh.Disconnect(false)
	if err != nil {
		return
	}

	// define table
	tr := true
	fl := false
	err = r.Init(dh, `frt.freight.init`, []Column{
		{
			Name: `freight_key`,
			Type: `nchar(38)`,
			Null: &tr,
		},
		{
			Name: `ref_address`,
			Type: `nvarchar(400)`,
			Null: &fl,
		},
		{
			Name: `ref_amount`,
			Type: `decimal(18,3)`,
			Null: &fl,
		},
	}, []string{
		`freight_key`,
	})

	if err != nil {
		log.Printf("Error found: %s\r\n", err.Error())
	}

	data := `{
				"freight_key": "1QuIyNTdgyCQtxHeVuHxROrtw3B",
				"ref_address": "21 Jump St. Quezon City",
				"ref_amount": 1.10
			 }`

	err = r.Delete(dh, `frt.freight.deleted`, []byte(data))
	if err != nil {
		log.Printf("Error found: %s\r\n", err.Error())
	}
}

func TestLoadReplicator(t *testing.T) {

	config, err := cfg.LoadConfig(`./config.json`)
	if err != nil {
		log.Fatal("Configuration file not found!")
	}

	// connect to database
	dh := datahelper.NewDataHelper(config)
	_, err = dh.Connect()
	defer dh.Disconnect(false)
	if err != nil {
		return
	}

	rp, err := LoadReplicator(dh, `./replicator.json`, false)
	log.Printf("Replicators: %v", rp)
}

func TestTimeConversion(t *testing.T) {
	// var ifc interface{}
	// ifc = `2019-09-11T00:00:00Z`
	// log.Println(anytstr(ifc))

	// ifc = `2019-09-18T06:24:15.2669612Z`
	// log.Println(anytstr(ifc))

	// ifc = `TEST0001`
	// log.Println(anytstr(ifc))

	// ifc = 3
	// log.Println(anytstr(ifc))
	r := NewReplicator(false)
	log.Println(r.rawtstr([]byte(`2019-09-11T00:00:00.000Z`)))
	log.Println(r.rawtstr([]byte(`2019-09-18T06:24:15.2669612Z`)))
	log.Println(r.rawtstr([]byte(`150,405.2669612Z`)))
}
