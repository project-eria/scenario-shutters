# ERIA Project - Scenario for opening / closing shutters

## Configuration file (scenario-shutters.json)
````
{
    "Lat": <latitude for house>,
    "Long": <longitude for house>,
    "OffsetOpen": 60,
    "OffsetClose": 90,
    "Devices": {
        "sejour": "<xAAL address>",
        "cuisine": "<xAAL address>",
        "salleAManger": "<xAAL address>",
        "bureau": "<xAAL address>",
        "ch1": "<xAAL address>",
        "ch2": "<xAAL address>"
    },
    "Schedules": [
        {
            "Days" : ["Saturday", "Sunday", "Wednesday"],
            "Open" : [
                {
                    "Time": "sunrise",
                    "Shutters" : ["sejour", "cuisine", "bureau", "salleAManger"]
                },
                {
                    "Time": "9:00",
                    "Shutters" : ["ch2"]
                },
                {
                    "Time" : "10:00",
                    "Shutters" : ["ch1"]
                }
            ],
            "Close" : [
                {
                    "Time": "sunset",
                    "Shutters" : ["sejour", "cuisine", "bureau", "ch1", "ch2"]
                },
                {
                    "Time": "23:00",
                    "Shutters" : ["salleAManger"]
                }
            ]
        },
        {
            "Days" : ["Monday", "Tuesday", "Thursday", "Friday"],
            "Open" : [
                {
                    "Time" : "sunrise",
                    "Shutters" : ["sejour", "cuisine", "bureau", "ch2", "salleAManger"]
                },
                {
                    "Time" : "10:00",
                    "Shutters" : ["ch1"]
                }
            ],
            "Close" : [
                {
                    "Shutters" : ["sejour", "cuisine", "bureau", "ch1", "ch2"],
                    "Time": "sunset"
                },
                {
                    "Shutters" : ["salleAManger"],
                    "Time": "23:00"
                }
            ]
        }
    ]
}
````