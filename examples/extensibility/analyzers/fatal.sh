#!/bin/bash
cat <<EOF
{
    "reports": [
        {
            "migration": {},
            "text": "Analyzer found ILLEGAL SQL statements",
            "diagnostics": [
                {
                    "lineNumber": -1,
                    "linePosition": -1,
                    "text": "PROHIBITED SQL statement found",
                    "code": "FAKE-002",
                    "level": "FATAL"
                }
            ],
            "actions": []
        }
    ]
}
EOF
