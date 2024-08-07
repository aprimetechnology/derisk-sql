#!/bin/bash
cat <<EOF
{
    "reports": [
        {
            "migration": {},
            "text": "Analyzer found a warning!",
            "diagnostics": [
                {
                    "lineNumber": -1,
                    "linePosition": -1,
                    "text": "Valid but discouraged SQL statement found",
                    "code": "FAKE-001",
                    "level": "WARNING"
                }
            ],
            "actions": []
        }
    ]
}
EOF
