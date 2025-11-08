## Test

```beforeAll
rm -f .local.db
```

```afterAll
rm -f .local.db
```

### Setup test data for advanced find operations

Create comprehensive test data for advanced find scenarios.

```execute
aux4 repository write products --id prod1 --data '{"name":"Laptop Pro","price":1299.99,"category":"electronics","inStock":true,"rating":4.8,"tags":["computer","portable","work"],"specs":{"cpu":"Intel i7","ram":"16GB","storage":"512GB SSD"}}' --metadata '{"vendor":"TechCorp","warehouse":"US-East"}'
aux4 repository write products --id prod2 --data '{"name":"Gaming Mouse","price":79.99,"category":"electronics","inStock":false,"rating":4.5,"tags":["gaming","mouse","rgb"],"specs":{"dpi":"16000","buttons":8,"wireless":true}}' --metadata '{"vendor":"GameGear","warehouse":"US-West"}'
aux4 repository write products --id prod3 --data '{"name":"Office Chair","price":249.50,"category":"furniture","inStock":true,"rating":4.2,"tags":["chair","ergonomic","office"],"specs":{"material":"leather","adjustable":true,"warranty":"5 years"}}' --metadata '{"vendor":"FurniCo","warehouse":"EU-Central"}'
aux4 repository write products --id prod4 --data '{"name":"Wireless Headphones","price":199.00,"category":"electronics","inStock":true,"rating":4.7,"tags":["audio","wireless","noise-canceling"],"specs":{"battery":"30 hours","noise_reduction":true,"color":"black"}}' --metadata '{"vendor":"AudioTech","warehouse":"US-East"}'
aux4 repository write products --id prod5 --data '{"name":"Standing Desk","price":549.00,"category":"furniture","inStock":false,"rating":4.3,"tags":["desk","standing","adjustable"],"specs":{"height_range":"28-48 inches","material":"bamboo","weight_capacity":"150 lbs"}}' --metadata '{"vendor":"FurniCo","warehouse":"EU-Central"}'
```

### Find using less than or equal operator

Find products with price less than or equal to $250.

```execute
aux4 repository find products --expr "price <= 250"
```

```expect:partial
[
  {
    "id": "prod2",
    "name": "Gaming Mouse",
    "price": 79.99,
    "category": "electronics",
    "inStock": false,
    "rating": 4.5,
    "tags": [
      "gaming",
      "mouse",
      "rgb"
    ],
    "specs": {
      "dpi": "16000",
      "buttons": 8,
      "wireless": true
    },
    "$metadata": {
      "vendor": "GameGear",
      "warehouse": "US-West",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "prod3",
    "name": "Office Chair",
    "price": 249.50,
    "category": "furniture",
    "inStock": true,
    "rating": 4.2,
    "tags": [
      "chair",
      "ergonomic",
      "office"
    ],
    "specs": {
      "material": "leather",
      "adjustable": true,
      "warranty": "5 years"
    },
    "$metadata": {
      "vendor": "FurniCo",
      "warehouse": "EU-Central",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "prod4",
    "name": "Wireless Headphones",
    "price": 199.00,
    "category": "electronics",
    "inStock": true,
    "rating": 4.7,
    "tags": [
      "audio",
      "wireless",
      "noise-canceling"
    ],
    "specs": {
      "battery": "30 hours",
      "noise_reduction": true,
      "color": "black"
    },
    "$metadata": {
      "vendor": "AudioTech",
      "warehouse": "US-East",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Find using greater than or equal operator

Find products with rating greater than or equal to 4.5.

```execute
aux4 repository find products --expr "rating >= 4.5"
```

```expect:partial
[
  {
    "id": "prod1",
    "name": "Laptop Pro",
    "price": 1299.99,
    "category": "electronics",
    "inStock": true,
    "rating": 4.8,
    "tags": [
      "computer",
      "portable",
      "work"
    ],
    "specs": {
      "cpu": "Intel i7",
      "ram": "16GB",
      "storage": "512GB SSD"
    },
    "$metadata": {
      "vendor": "TechCorp",
      "warehouse": "US-East",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "prod2",
    "name": "Gaming Mouse",
    "price": 79.99,
    "category": "electronics",
    "inStock": false,
    "rating": 4.5,
    "tags": [
      "gaming",
      "mouse",
      "rgb"
    ],
    "specs": {
      "dpi": "16000",
      "buttons": 8,
      "wireless": true
    },
    "$metadata": {
      "vendor": "GameGear",
      "warehouse": "US-West",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "prod4",
    "name": "Wireless Headphones",
    "price": 199.00,
    "category": "electronics",
    "inStock": true,
    "rating": 4.7,
    "tags": [
      "audio",
      "wireless",
      "noise-canceling"
    ],
    "specs": {
      "battery": "30 hours",
      "noise_reduction": true,
      "color": "black"
    },
    "$metadata": {
      "vendor": "AudioTech",
      "warehouse": "US-East",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Find using not equal operator variants

Test both != and <> operators for inequality.

```execute
aux4 repository find products --expr "category != 'electronics'"
```

```expect:partial
[
  {
    "id": "prod3",
    "name": "Office Chair",
    "price": 249.50,
    "category": "furniture",
    "inStock": true,
    "rating": 4.2,
    "tags": [
      "chair",
      "ergonomic",
      "office"
    ],
    "specs": {
      "material": "leather",
      "adjustable": true,
      "warranty": "5 years"
    },
    "$metadata": {
      "vendor": "FurniCo",
      "warehouse": "EU-Central",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "prod5",
    "name": "Standing Desk",
    "price": 549.00,
    "category": "furniture",
    "inStock": false,
    "rating": 4.3,
    "tags": [
      "desk",
      "standing",
      "adjustable"
    ],
    "specs": {
      "height_range": "28-48 inches",
      "material": "bamboo",
      "weight_capacity": "150 lbs"
    },
    "$metadata": {
      "vendor": "FurniCo",
      "warehouse": "EU-Central",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

```execute
aux4 repository find products --expr "category <> 'furniture'"
```

```expect:partial
[
  {
    "id": "prod1",
    "name": "Laptop Pro",
    "price": 1299.99,
    "category": "electronics",
    "inStock": true,
    "rating": 4.8,
    "tags": [
      "computer",
      "portable",
      "work"
    ],
    "specs": {
      "cpu": "Intel i7",
      "ram": "16GB",
      "storage": "512GB SSD"
    },
    "$metadata": {
      "vendor": "TechCorp",
      "warehouse": "US-East",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "prod2",
    "name": "Gaming Mouse",
    "price": 79.99,
    "category": "electronics",
    "inStock": false,
    "rating": 4.5,
    "tags": [
      "gaming",
      "mouse",
      "rgb"
    ],
    "specs": {
      "dpi": "16000",
      "buttons": 8,
      "wireless": true
    },
    "$metadata": {
      "vendor": "GameGear",
      "warehouse": "US-West",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "prod4",
    "name": "Wireless Headphones",
    "price": 199.00,
    "category": "electronics",
    "inStock": true,
    "rating": 4.7,
    "tags": [
      "audio",
      "wireless",
      "noise-canceling"
    ],
    "specs": {
      "battery": "30 hours",
      "noise_reduction": true,
      "color": "black"
    },
    "$metadata": {
      "vendor": "AudioTech",
      "warehouse": "US-East",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Find using NOT LIKE operator

Find products where name does not contain specific patterns.

```execute
aux4 repository find products --expr "name not like '%Gaming%'"
```

```expect:partial
[
  {
    "id": "prod1",
    "name": "Laptop Pro",
    "price": 1299.99,
    "category": "electronics",
    "inStock": true,
    "rating": 4.8,
    "tags": [
      "computer",
      "portable",
      "work"
    ],
    "specs": {
      "cpu": "Intel i7",
      "ram": "16GB",
      "storage": "512GB SSD"
    },
    "$metadata": {
      "vendor": "TechCorp",
      "warehouse": "US-East",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "prod3",
    "name": "Office Chair",
    "price": 249.50,
    "category": "furniture",
    "inStock": true,
    "rating": 4.2,
    "tags": [
      "chair",
      "ergonomic",
      "office"
    ],
    "specs": {
      "material": "leather",
      "adjustable": true,
      "warranty": "5 years"
    },
    "$metadata": {
      "vendor": "FurniCo",
      "warehouse": "EU-Central",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "prod4",
    "name": "Wireless Headphones",
    "price": 199.00,
    "category": "electronics",
    "inStock": true,
    "rating": 4.7,
    "tags": [
      "audio",
      "wireless",
      "noise-canceling"
    ],
    "specs": {
      "battery": "30 hours",
      "noise_reduction": true,
      "color": "black"
    },
    "$metadata": {
      "vendor": "AudioTech",
      "warehouse": "US-East",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "prod5",
    "name": "Standing Desk",
    "price": 549.00,
    "category": "furniture",
    "inStock": false,
    "rating": 4.3,
    "tags": [
      "desk",
      "standing",
      "adjustable"
    ],
    "specs": {
      "height_range": "28-48 inches",
      "material": "bamboo",
      "weight_capacity": "150 lbs"
    },
    "$metadata": {
      "vendor": "FurniCo",
      "warehouse": "EU-Central",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Find using OR conditions

First add some simple test data for OR conditions.

```execute
aux4 repository write simple_items --id item1 --data '{"category":"electronics","price":100}' --metadata '{}'
```

```execute
aux4 repository write simple_items --id item2 --data '{"category":"furniture","price":200}' --metadata '{}'
```

```execute
aux4 repository write simple_items --id item3 --data '{"category":"electronics","price":300}' --metadata '{}'
```

Find items that are either electronics OR have price less than 250.

```execute
aux4 repository find simple_items --expr "category = 'electronics' or price < 250"
```

```expect:partial
[
  {
    "id": "item1",
    "category": "electronics",
    "price": 100,
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "item2",
    "category": "furniture",
    "price": 200,
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "item3",
    "category": "electronics",
    "price": 300,
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Find using complex compound expressions with parentheses

Find products that are (electronics AND in stock) OR (furniture AND price > 300).

```execute
aux4 repository find products --expr "(category = 'electronics' and inStock = 1) or (category = 'furniture' and price > 300)"
```

```expect:partial
[
  {
    "id": "prod1",
    "name": "Laptop Pro",
    **
    "category": "electronics",
    "inStock": true,
    **
  },
  {
    "id": "prod4",
    "name": "Wireless Headphones",
    **
    "category": "electronics",
    "inStock": true,
    **
  },
  {
    "id": "prod5",
    "name": "Standing Desk",
    **
    "category": "furniture",
    "inStock": false,
    **
  }
]
```

### Find using nested JSON path queries

Add simple test data for JSON path queries.

```execute
aux4 repository write devices --id device1 --data '{"specs":{"wireless":true},"name":"Mouse"}' --metadata '{}'
```

```execute
aux4 repository write devices --id device2 --data '{"specs":{"wireless":false},"name":"Keyboard"}' --metadata '{}'
```

Find devices with wireless specifications.

```execute
aux4 repository find devices --expr "json_extract(data, '$.specs.wireless') = 1"
```

```expect:partial
[
  {
    "id": "device1",
    "specs": {
      "wireless": true
    },
    "name": "Mouse",
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Find using LIKE with multiple wildcards

Add simple test data for wildcard matching.

```execute
aux4 repository write names --id n1 --data '{"name":"Apple"}' --metadata '{}'
```

```execute
aux4 repository write names --id n2 --data '{"name":"Orange"}' --metadata '{}'
```

```execute
aux4 repository write names --id n3 --data '{"name":"Grape"}' --metadata '{}'
```

Find names containing both 'a' and 'e'.

```execute
aux4 repository find names --expr "name like '%a%' and name like '%e%'"
```

```expect:partial
[
  {
    "id": "n1",
    "name": "Apple",
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "n2",
    "name": "Orange",
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "n3",
    "name": "Grape",
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Find with case-insensitive LIKE patterns

Test LIKE patterns with different cases.

```execute
aux4 repository find products --expr "name like '%GAMING%' or name like '%office%'"
```

```expect:partial
[
  {
    "id": "prod2",
    "name": "Gaming Mouse",
    **
  },
  {
    "id": "prod3",
    "name": "Office Chair",
    **
  }
]
```

### Find using numeric ranges

Add simple test data for numeric ranges.

```execute
aux4 repository write prices --id p1 --data '{"price":50}' --metadata '{}'
```

```execute
aux4 repository write prices --id p2 --data '{"price":150}' --metadata '{}'
```

```execute
aux4 repository write prices --id p3 --data '{"price":250}' --metadata '{}'
```

Find items with price between 100 and 200.

```execute
aux4 repository find prices --expr "price >= 100 and price <= 200"
```

```expect:partial
[
  {
    "id": "p2",
    "price": 150,
    "$metadata": {
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```

### Find using boolean false values

Find products that are not in stock (boolean false).

```execute
aux4 repository find products --expr "inStock = 0"
```

```expect:partial
[
  {
    "id": "prod2",
    "name": "Gaming Mouse",
    **
    "inStock": false,
    **
  },
  {
    "id": "prod5",
    "name": "Standing Desk",
    **
    "inStock": false,
    **
  }
]
```

### Find with multiple nested JSON conditions

Find products with specific nested specifications using multiple JSON path extractions.

```execute
aux4 repository find products --expr "json_extract(data, '$.specs.material') = 'leather' or json_extract(data, '$.specs.battery') like '%30%'"
```

```expect:partial
[
  {
    "id": "prod3",
    "name": "Office Chair",
    **
    "specs": {
      "material": "leather",
      **
    },
    **
  },
  {
    "id": "prod4",
    "name": "Wireless Headphones",
    **
    "specs": {
      "battery": "30 hours",
      **
    },
    **
  }
]
```

### Find using string length and pattern combinations

Find products where category length and pattern match specific criteria.

```execute
aux4 repository find products --expr "length(json_extract(data, '$.category')) > 10 and json_extract(data, '$.category') like '%elect%'"
```

```expect:partial
[
  {
    "id": "prod1",
    "name": "Laptop Pro",
    "price": 1299.99,
    "category": "electronics",
    "inStock": true,
    "rating": 4.8,
    "tags": [
      "computer",
      "portable",
      "work"
    ],
    "specs": {
      "cpu": "Intel i7",
      "ram": "16GB",
      "storage": "512GB SSD"
    },
    "$metadata": {
      "vendor": "TechCorp",
      "warehouse": "US-East",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "prod2",
    "name": "Gaming Mouse",
    "price": 79.99,
    "category": "electronics",
    "inStock": false,
    "rating": 4.5,
    "tags": [
      "gaming",
      "mouse",
      "rgb"
    ],
    "specs": {
      "dpi": "16000",
      "buttons": 8,
      "wireless": true
    },
    "$metadata": {
      "vendor": "GameGear",
      "warehouse": "US-West",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  },
  {
    "id": "prod4",
    "name": "Wireless Headphones",
    "price": 199.00,
    "category": "electronics",
    "inStock": true,
    "rating": 4.7,
    "tags": [
      "audio",
      "wireless",
      "noise-canceling"
    ],
    "specs": {
      "battery": "30 hours",
      "noise_reduction": true,
      "color": "black"
    },
    "$metadata": {
      "vendor": "AudioTech",
      "warehouse": "US-East",
      "createdAt": "*?",
      "updatedAt": "*?"
    }
  }
]
```
