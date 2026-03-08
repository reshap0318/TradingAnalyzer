#!/bin/bash

# --- Konfigurasi ---
BASE_URL="http://trading-bot:8000"
LOG_FILE="curl_activity.log"

# Data Login
USERNAME="admin"
PASSWORD="tLoKGbW59AFwqb%Z"

SYMBOLS=(
  "BTCUSDT" "ETHUSDT" "BNBUSDT" "SOLUSDT" "XRPUSDT" 
  "ADAUSDT" "DOGEUSDT" "TRXUSDT" "LINKUSDT" "AVAXUSDT" 
  "SUIUSDT" "LTCUSDT" "NEARUSDT" "UNIUSDT" "FETUSDT"
)

# --- Fungsi Login ---
do_login() {
    echo "[$(date "+%Y-%m-%d %H:%M:%S")] [AUTH] Attempting login..." >> $LOG_FILE
    
    # Hit API Login
    LOGIN_RESPONSE=$(curl -s --location "$BASE_URL/api/auth/login" \
    --header 'Content-Type: application/json' \
    --data "{
        \"username\": \"$USERNAME\",
        \"password\": \"$PASSWORD\"
    }")

    # Cara mengambil token:
    # 1. Cari kata "token"
    # 2. Ambil string di antara tanda kutip berikutnya
    NEW_TOKEN=$(echo $LOGIN_RESPONSE | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')

    if [ -z "$NEW_TOKEN" ]; then
        echo "[$(date "+%Y-%m-%d %H:%M:%S")] [AUTH] Login FAILED. Response: $LOGIN_RESPONSE" >> $LOG_FILE
        return 1
    else
        echo "[$(date "+%Y-%m-%d %H:%M:%S")] [AUTH] Login SUCCESS. Token updated." >> $LOG_FILE
        TOKEN=$NEW_TOKEN
        return 0
    fi
}

# --- Fungsi untuk Watcher (Jalan setiap 60 detik) ---
run_watcher() {
    while true; do
        if [ ! -z "$TOKEN" ]; then
            current_time=$(date "+%Y-%m-%d %H:%M:%S")
            response_watcher=$(curl -s --location --request POST "$BASE_URL/api/signal/watcher" \
            --header "Authorization: Bearer $TOKEN")
            
            echo "[$current_time] [WATCHER] Response: $response_watcher" >> $LOG_FILE
        fi
        sleep 60
    done
}

# --- Main Logic ---
echo "=== Bot Started: $(date) ===" >> $LOG_FILE

# Jalankan login pertama kali
do_login
if [ $? -ne 0 ]; then
    echo "Initial login failed. Exiting..." >> $LOG_FILE
    exit 1
fi

# Jalankan watcher di background setelah dapat token
run_watcher &

while true; do
    current_time=$(date "+%Y-%m-%d %H:%M:%S")
    echo "[$current_time] [SIGNAL] Starting batch for 15 symbols..." >> $LOG_FILE
    
    for SYM in "${SYMBOLS[@]}"; do
        ts=$(date "+%Y-%m-%d %H:%M:%S")
        
        response_signal=$(curl -s --location "$BASE_URL/api/signal/create" \
        --header 'Content-Type: application/json' \
        --header "Authorization: Bearer $TOKEN" \
        --data "{ \"symbol\": \"$SYM\" }")
        
        # Cek jika token expired (biasanya return 401 Unauthorized)
        if [[ $response_signal == *"Unauthenticated"* ]] || [[ $response_signal == *"Unauthorized"* ]]; then
            echo "[$ts] Token expired, relogging..." >> $LOG_FILE
            do_login
            # Ulangi hit koin yang sama dengan token baru
            response_signal=$(curl -s --location "$BASE_URL/api/signal/create" \
            --header 'Content-Type: application/json' \
            --header "Authorization: Bearer $TOKEN" \
            --data "{ \"symbol\": \"$SYM\" }")
        fi

        echo "[$ts] Symbol: $SYM | Response: $response_signal" >> $LOG_FILE
        sleep 29
    done
    
    echo "[$(date "+%Y-%m-%d %H:%M:%S")] [SIGNAL] Batch finished. Waiting for next cycle..." >> $LOG_FILE
    sleep 900
done