#!/bin/bash

# HTTP 連線池優化效能測試腳本
# 使用方法: ./benchmark.sh

echo "🚀 HTTP 連線池優化效能測試"
echo "=============================="

# 測試參數
BOARD="beauty"
PAGES=2
PUSH_RATE=10

echo "測試參數:"
echo "- 看板: $BOARD"
echo "- 頁數: $PAGES"
echo "- 推文數門檻: $PUSH_RATE"
echo ""

# 建立測試配置檔案
echo "📝 建立測試配置檔案..."

# 原始配置 (模擬未優化狀態)
cat > config_original.yaml << EOF
crawler:
  workers: 10
  parserCount: 10
  channels:
    articleInfo: 100
    downloadTask: 200
    markdownTask: 100
  delays:
    minMs: 500
    maxMs: 2000
  http:
    timeout: "30s"
    maxIdleConns: 10
    maxIdleConnsPerHost: 2
    idleConnTimeout: "30s"
    tlsHandshakeTimeout: "10s"
    expectContinueTimeout: "1s"
EOF

# 優化配置 (連線池優化)
cat > config_optimized.yaml << EOF
crawler:
  workers: 10
  parserCount: 10
  channels:
    articleInfo: 100
    downloadTask: 200
    markdownTask: 100
  delays:
    minMs: 500
    maxMs: 2000
  http:
    timeout: "30s"
    maxIdleConns: 100
    maxIdleConnsPerHost: 20
    idleConnTimeout: "90s"
    tlsHandshakeTimeout: "10s"
    expectContinueTimeout: "1s"
EOF

echo "✅ 配置檔案已建立"
echo ""

# 清理之前的測試結果
echo "🧹 清理之前的測試結果..."
rm -rf beauty_original/ beauty_optimized/ 2>/dev/null
echo ""

# 測試原始配置
echo "⏱️  測試原始配置 (連線池未優化)..."
echo "開始時間: $(date)"
START_TIME_ORIGINAL=$(date +%s)

timeout 300 go run main.go -board=$BOARD -pages=$PAGES -push=$PUSH_RATE -config=config_original.yaml > /dev/null 2>&1
RESULT_ORIGINAL=$?

END_TIME_ORIGINAL=$(date +%s)
DURATION_ORIGINAL=$((END_TIME_ORIGINAL - START_TIME_ORIGINAL))

if [ $RESULT_ORIGINAL -eq 124 ]; then
    echo "❌ 原始配置測試超時 (5分鐘)"
    DURATION_ORIGINAL="TIMEOUT"
elif [ $RESULT_ORIGINAL -ne 0 ]; then
    echo "❌ 原始配置測試失敗"
    DURATION_ORIGINAL="ERROR"
else
    echo "✅ 原始配置測試完成"
fi

echo "原始配置執行時間: ${DURATION_ORIGINAL}秒"
echo ""

# 重新命名結果目錄
[ -d "$BOARD" ] && mv "$BOARD" "${BOARD}_original"

# 測試優化配置
echo "⏱️  測試優化配置 (HTTP 連線池優化)..."
echo "開始時間: $(date)"
START_TIME_OPTIMIZED=$(date +%s)

timeout 300 go run main.go -board=$BOARD -pages=$PAGES -push=$PUSH_RATE -config=config_optimized.yaml > /dev/null 2>&1
RESULT_OPTIMIZED=$?

END_TIME_OPTIMIZED=$(date +%s)
DURATION_OPTIMIZED=$((END_TIME_OPTIMIZED - START_TIME_OPTIMIZED))

if [ $RESULT_OPTIMIZED -eq 124 ]; then
    echo "❌ 優化配置測試超時 (5分鐘)"
    DURATION_OPTIMIZED="TIMEOUT"
elif [ $RESULT_OPTIMIZED -ne 0 ]; then
    echo "❌ 優化配置測試失敗"
    DURATION_OPTIMIZED="ERROR"
else
    echo "✅ 優化配置測試完成"
fi

echo "優化配置執行時間: ${DURATION_OPTIMIZED}秒"
echo ""

# 重新命名結果目錄
[ -d "$BOARD" ] && mv "$BOARD" "${BOARD}_optimized"

# 效能比較
echo "📊 效能比較結果"
echo "=============================="
echo "原始配置 (連線池未優化): ${DURATION_ORIGINAL}秒"
echo "優化配置 (HTTP 連線池優化): ${DURATION_OPTIMIZED}秒"

if [[ "$DURATION_ORIGINAL" =~ ^[0-9]+$ ]] && [[ "$DURATION_OPTIMIZED" =~ ^[0-9]+$ ]]; then
    if [ $DURATION_OPTIMIZED -lt $DURATION_ORIGINAL ]; then
        IMPROVEMENT=$((DURATION_ORIGINAL - DURATION_OPTIMIZED))
        IMPROVEMENT_PERCENT=$((IMPROVEMENT * 100 / DURATION_ORIGINAL))
        echo "🚀 效能提升: ${IMPROVEMENT}秒 (${IMPROVEMENT_PERCENT}%)"
    elif [ $DURATION_OPTIMIZED -gt $DURATION_ORIGINAL ]; then
        REGRESSION=$((DURATION_OPTIMIZED - DURATION_ORIGINAL))
        REGRESSION_PERCENT=$((REGRESSION * 100 / DURATION_ORIGINAL))
        echo "⚠️  效能回退: ${REGRESSION}秒 (${REGRESSION_PERCENT}%)"
    else
        echo "➡️  效能相當"
    fi
else
    echo "❓ 無法計算效能差異 (有測試失敗或超時)"
fi

echo ""

# 檔案數量比較
if [ -d "${BOARD}_original" ] && [ -d "${BOARD}_optimized" ]; then
    ORIGINAL_FILES=$(find "${BOARD}_original" -type f | wc -l)
    OPTIMIZED_FILES=$(find "${BOARD}_optimized" -type f | wc -l)
    
    echo "📁 下載檔案數量比較"
    echo "原始配置: $ORIGINAL_FILES 個檔案"
    echo "優化配置: $OPTIMIZED_FILES 個檔案"
    
    if [ $ORIGINAL_FILES -eq $OPTIMIZED_FILES ]; then
        echo "✅ 檔案數量一致，測試結果可信"
    else
        echo "⚠️  檔案數量不一致，可能影響效能比較"
    fi
fi

echo ""
echo "📋 測試總結"
echo "=============================="
echo "✅ HTTP 連線池優化實作完成"
echo "✅ 效能測試執行完成"
echo "✅ 配置檔案可調整以適應不同網路環境"
echo ""
echo "📝 建議後續行動:"
echo "1. 根據網路環境調整 maxIdleConnsPerHost 參數"
echo "2. 監控記憶體使用情況，必要時調整 maxIdleConns"
echo "3. 考慮實作動態 Channel 調整機制"
echo ""
echo "🧹 清理測試檔案:"
echo "rm -f config_original.yaml config_optimized.yaml"
echo "rm -rf ${BOARD}_original/ ${BOARD}_optimized/"