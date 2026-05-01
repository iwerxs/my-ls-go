#!/bin/bash

TARGET="/usr/bin"

echo "------------------------------------------------"
echo "AUDIT Q7: Testing -l on $TARGET"
echo ""

# Run commands
ls -l "$TARGET" > expected.txt
./my-ls "$TARGET" > actual.txt

# Clean raw outputs (remove ACL '+')
sed 's/+//' expected.txt > expected_clean.txt
sed 's/+//' actual.txt > actual_clean.txt

echo "=== STRICT DIFF (raw comparison) ==="
if diff expected_clean.txt actual_clean.txt > /dev/null; then
    echo "RESULT: ✅ STRICT MATCH"
else
    echo "RESULT: ❌ STRICT DIFFERENCE"
    diff -u expected_clean.txt actual_clean.txt | head -n 20
fi

echo ""

# --- STRUCTURE NORMALISATION ---
# Remove volatile fields (size + date)
normalize() {
    awk '
    BEGIN { OFS=" " }
    {
        if ($1 == "total") {
            print "total"
        } else {
            # Keep:
            # $1 = mode
            # $2 = links
            # $3 = owner
            # $4 = group
            # $NF = filename (and possible symlink arrow)
            
            name=""
            for (i=9; i<=NF; i++) {
                name = name $i " "
            }

            print $1, $2, $3, $4, name
        }
    }'
}

normalize < expected_clean.txt > expected_stable.txt
normalize < actual_clean.txt > actual_stable.txt

echo "=== STRUCTURE DIFF (ignoring size + timestamps + ACL) ==="
if diff expected_stable.txt actual_stable.txt > /dev/null; then
    echo "RESULT: ✅ STRUCTURE MATCHES"
    echo "PASS: Output is consistent with system ls"
else
    echo "RESULT: ❌ STRUCTURE MISMATCH"
    echo ""
    echo "Showing first differences:"
    diff -u expected_stable.txt actual_stable.txt | head -n 40
fi

echo ""
echo "=== PERFORMANCE TEST ==="
time ./my-ls "$TARGET" > /dev/null

echo "------------------------------------------------"