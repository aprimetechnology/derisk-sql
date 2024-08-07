#!/bin/bash

# copy stdin to a local file
tempFile=$(mktemp)
cat - > ${tempFile}

# print the start of the reports array
cat <<EOF
{
  "reports": [
EOF

# loop through all migration files
numMigrations=$(( $(cat ${tempFile} | jq '.migrations | length') - 1 ))
for i in $(seq 0 ${numMigrations});
do
    # grep the migration file contents for the literal phrase 'DROP TABLE'
    if cat ${tempFile} | jq ".migrations[${i}].down" | grep -q "DROP TABLE";
    then
        # print a report for each offending migration
        cat <<EOF
    {
      "migration": {
$(cat ${tempFile} | jq ."migrations[${i}]" -M | sed 's/^/      /g' | tail -n +2),
      "text": "Migration failed this analyzer!",
      "diagnostics": [
        {
          "lineNumber": -1,
          "linePosition": -1,
          "text": "Found illegal DROP TABLE statement!!",
          "code": "DROP-123",
          "level": "FATAL"
        }
      ],
      "actions": []
    }
EOF
    fi
done

# print the end of the reports array
cat <<EOF
  ]
}
EOF
