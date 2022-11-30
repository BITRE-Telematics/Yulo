package ys

import (
	"fmt"
	//"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// upload_veh data adds vehicle firm and type to the database
func upload_veh_data(id string, veh_type string, firm string) {
	fmt.Println("Adding vehicle info %s %s %s", id, veh_type, firm)
	session := Db.NewSession(Sesh_config)

	defer session.Close()
	statement := `
				MERGE(a:Asset{id: $ID})
				WITH a
				//WHERE a.firm = 'Unknown' OR a.firm IS NULL
				SET a.type = $VEH_TYPE
				SET a.firm = $FIRM
				return a.type as type, a.firm as firm
				`
	parameters := map[string]interface{}{
		"ID":       id,
		"VEH_TYPE": veh_type,
		"FIRM":     firm,
	}
	//fmt.Println(parameters)
	result, err := session.Run(statement, parameters)
	if err != nil {
		fmt.Print("Add veh data error")
	}
	if result.Err() != nil {
		fmt.Print("Add veh data error")
		fmt.Println(result.Err())
	}
	label := "Vehicle"
	if veh_type == "Trailer" {
		label = "Trailer"
	}
	//we can potentially end up with mislabeled assets where trailers are initially unknown.
	statement2 := fmt.Sprintf(`
		MERGE(a:Asset{id: $ID}) 
		WITH a 
		WHERE SIZE(LABELS(a)) < 2 
		SET a:%s RETURN a.id
		`,
		label)
	parameters2 := map[string]interface{}{
		"ID": id,
	}

	result2, err := session.Run(statement2, parameters2)
	if err != nil {
		fmt.Println("Add veh data error")
		fmt.Println(err)
	}
	if result2 != nil {
		if result2.Err() != nil {
			fmt.Println("Add veh data labels error")
			fmt.Println(result.Err())
		}
	}

}
