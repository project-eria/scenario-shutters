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
            "Sets" : [
                {
                    "Shutters" : ["sejour" "cuisine", "bureau"],
                    "OpenTime" : "sunrise",
                    "CloseTime": "sunset"
                },
                {
                    "Shutters" : ["salleAManger"],
                    "OpenTime" : "sunrise",
                    "CloseTime": "23:00"
                },
                {
                    "Shutters" : ["ch1", "ch2"],
                    "OpenTime" : "9:00",
                    "CloseTime": "sunset"
                }
            ]
        },
        {
            "Days" : ["Monday", "Tuesday", "Thursday", "Friday"],
            "Sets" : [
                {
                    "Shutters" : ["sejour", "cuisine", "bureau", "ch2"],
                    "OpenTime" : "sunrise",
                    "CloseTime": "sunset"
                },
                {
                    "Shutters" : ["salleAManger"],
                    "OpenTime" : "sunrise",
                    "CloseTime": "23:00"
                },
                {
                    "Shutters" : ["ch1"],
                    "OpenTime" : "10:00",
                    "CloseTime": "sunset"
                }
            ]
        }
    ]
}
````