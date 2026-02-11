table "cards" {
  schema = schema.public
  column "id" {
    null = false
    type = text
  }
  column "artist" {
    null = true
    type = text
  }
  column "asciiname" {
    null = true
    type = text
  }
  column "availability" {
    null = true
    type = text
  }
  column "bordercolor" {
    null = true
    type = text
  }
  column "cardkingdomid" {
    null = true
    type = text
  }
  column "coloridentity" {
    null = true
    type = text
  }
  column "colorindicator" {
    null = true
    type = text
  }
  column "colors" {
    null = true
    type = text
  }
  column "convertedmanacost" {
    null = true
    type = text
  }
  column "faceconvertedmanacost" {
    null = true
    type = text
  }
  column "facemanavalue" {
    null = true
    type = text
  }
  column "flavorname" {
    null = true
    type = text
  }
  column "flavortext" {
    null = true
    type = text
  }
  column "keywords" {
    null = true
    type = text
  }
  column "mtgjsonv4id" {
    null = true
    type = text
  }
  column "name" {
    null = true
    type = text
  }
  column "number" {
    null = true
    type = text
  }
  column "originaltext" {
    null = true
    type = text
  }
  column "originaltype" {
    null = true
    type = text
  }
  column "power" {
    null = true
    type = text
  }
  column "scryfallid" {
    null = true
    type = text
  }
  column "scryfallillustrationid" {
    null = true
    type = text
  }
  column "scryfalloracleid" {
    null = true
    type = text
  }
  column "setcode" {
    null = true
    type = text
  }
  column "side" {
    null = true
    type = text
  }
  column "subtypes" {
    null = true
    type = text
  }
  column "supertypes" {
    null = true
    type = text
  }
  column "tcgplayerproductid" {
    null = true
    type = text
  }
  column "text" {
    null = true
    type = text
  }
  column "toughness" {
    null = true
    type = text
  }
  column "type" {
    null = true
    type = text
  }
  column "types" {
    null = true
    type = text
  }
  column "uuid" {
    null = false
    type = text
  }
  column "facename" {
    null = true
    type = text
  }
  primary_key {
    columns = [column.id]
  }
  index "cards_uuid_key" {
    unique  = true
    columns = [column.uuid]
  }
}
table "gamelog" {
  schema = schema.public
  column "id" {
    null = false
    type = bigint
  }
  column "game_id" {
    null = false
    type = text
  }
  column "payload" {
    null = true
    type = jsonb
  }
  column "eventtime" {
    null = true
    type = timestamp
  }
  primary_key {
    columns = [column.id]
  }
}
table "games" {
  schema = schema.public
  column "id" {
    null = true
    type = text
  }
  column "payload" {
    null = true
    type = jsonb
  }
  column "eventtime" {
    null = true
    type = timestamp
  }
  index "pkey_games" {
    unique  = true
    columns = [column.id]
  }
}
table "schema_migrations" {
  schema = schema.public
  column "version" {
    null = false
    type = bigint
  }
  column "dirty" {
    null = false
    type = boolean
  }
  primary_key {
    columns = [column.version]
  }
}
table "users" {
  schema = schema.public
  column "username" {
    null = true
    type = character_varying(255)
  }
  column "password" {
    null = true
    type = character_varying(255)
  }
  column "uuid" {
    null = true
    type = character_varying(255)
  }
  column "timestamp" {
    null    = true
    type    = timestamp
    default = sql("CURRENT_TIMESTAMP")
  }
  index "username_unique" {
    unique  = true
    columns = [column.username]
  }
  index "users_uuid_key" {
    unique  = true
    columns = [column.uuid]
  }
}
schema "public" {
}
