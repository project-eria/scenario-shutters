# ERIA Project - Scenario for opening / closing shutters

## Configuration file (scenario-shutters.json)
````
{
    "Lat": <latitude for house>,
    "Long": <longitude for house>,
    "Devices": {
        "sejour": "<xAAL address>",
        "cuisine": "<xAAL address>",
        "salleAManger": "<xAAL address>",
        "bureau": "<xAAL address>",
        "ch1": "<xAAL address>",
        "ch2": "<xAAL address>"
    },
    "Events" : [
        {
            "Label": "away",
            "Value": false
        },
        {
            "Label": "holiday",
            "Value": true
        }
    ],
    "Schedules": [
        {
            "Days" : ["Saturday", "Sunday", "Wednesday", "holiday"],
            "Open" : [
                {
                    "Time": "sunrise+10",
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
                    "Time": "sunset-5",
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
                    "Time" : "sunrise+30",
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
                    "Time": "sunset-5"
                },
                {
                    "Shutters" : ["salleAManger"],
                    "Time": "23:00"
                }
            ]
        },
        {
            "Days" : ["away"],
            "Open" : [
                {
                    "Time": "sunrise-20",
                    "Shutters" : ["sejour", "salleAManger", "cuisine", "bureau", "ch1", "ch2"]
                }
            ],
            "Close" : [
                {
                    "Time": "sunset+30",
                    "Shutters" : ["sejour", "salleAManger", "cuisine", "bureau", "ch1", "ch2"]
                }
            ]
        }
    ]
}
````