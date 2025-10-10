#!/bin/bash

# Export Service Discovery Documentation to HTML
# This script converts markdown documentation to HTML format

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/html"

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo "Exporting Service Discovery documentation to HTML..."

# Function to convert markdown to HTML using a simple approach
convert_md_to_html() {
    local input_file="$1"
    local output_file="$2"
    local title="$3"

    cat > "$output_file" <<'HTML_HEADER'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>TITLE_PLACEHOLDER</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background-color: white;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            color: #2c3e50;
            border-bottom: 3px solid #3498db;
            padding-bottom: 10px;
        }
        h2 {
            color: #34495e;
            margin-top: 30px;
            border-bottom: 2px solid #ecf0f1;
            padding-bottom: 8px;
        }
        h3 {
            color: #7f8c8d;
            margin-top: 20px;
        }
        code {
            background-color: #f8f8f8;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: 'Courier New', Courier, monospace;
            font-size: 0.9em;
        }
        pre {
            background-color: #2c3e50;
            color: #ecf0f1;
            padding: 15px;
            border-radius: 5px;
            overflow-x: auto;
        }
        pre code {
            background-color: transparent;
            color: #ecf0f1;
            padding: 0;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
        }
        table th {
            background-color: #3498db;
            color: white;
            padding: 12px;
            text-align: left;
        }
        table td {
            padding: 10px;
            border: 1px solid #ddd;
        }
        table tr:nth-child(even) {
            background-color: #f9f9f9;
        }
        blockquote {
            border-left: 4px solid #3498db;
            padding-left: 20px;
            margin: 20px 0;
            color: #555;
            background-color: #f0f8ff;
            padding: 10px 20px;
            border-radius: 4px;
        }
        a {
            color: #3498db;
            text-decoration: none;
        }
        a:hover {
            text-decoration: underline;
        }
        .toc {
            background-color: #f8f9fa;
            padding: 20px;
            border-radius: 5px;
            margin-bottom: 30px;
        }
        .toc ul {
            list-style-type: none;
            padding-left: 0;
        }
        .toc li {
            margin: 5px 0;
        }
        .note {
            background-color: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 15px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .warning {
            background-color: #f8d7da;
            border-left: 4px solid #dc3545;
            padding: 15px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .success {
            background-color: #d4edda;
            border-left: 4px solid #28a745;
            padding: 15px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .footer {
            margin-top: 50px;
            padding-top: 20px;
            border-top: 1px solid #ddd;
            text-align: center;
            color: #888;
        }
        @media print {
            body {
                background-color: white;
            }
            .container {
                box-shadow: none;
            }
        }
    </style>
</head>
<body>
    <div class="container">
HTML_HEADER

    sed -i "s/TITLE_PLACEHOLDER/$title/g" "$output_file"

    # Simple markdown to HTML conversion
    # This is a basic conversion - for production, use pandoc or similar
    sed -e 's/^# \(.*\)$/<h1>\1<\/h1>/' \
        -e 's/^## \(.*\)$/<h2>\1<\/h2>/' \
        -e 's/^### \(.*\)$/<h3>\1<\/h3>/' \
        -e 's/^#### \(.*\)$/<h4>\1<\/h4>/' \
        -e 's/^\* \(.*\)$/<li>\1<\/li>/' \
        -e 's/^- \(.*\)$/<li>\1<\/li>/' \
        -e 's/^[0-9]\+\. \(.*\)$/<li>\1<\/li>/' \
        -e 's/\*\*\([^*]*\)\*\*/<strong>\1<\/strong>/g' \
        -e 's/\*\([^*]*\)\*/<em>\1<\/em>/g' \
        -e 's/`\([^`]*\)`/<code>\1<\/code>/g' \
        "$input_file" >> "$output_file"

    cat >> "$output_file" <<'HTML_FOOTER'
    <div class="footer">
        <p>Generated on GENERATION_DATE</p>
        <p>&copy; 2025 HelixTrack Core - Service Discovery Documentation</p>
    </div>
    </div>
</body>
</html>
HTML_FOOTER

    sed -i "s/GENERATION_DATE/$(date '+%Y-%m-%d %H:%M:%S')/" "$output_file"

    echo "✓ Exported: $output_file"
}

# Export Technical Documentation
convert_md_to_html \
    "$SCRIPT_DIR/ServiceDiscovery_Technical.md" \
    "$OUTPUT_DIR/ServiceDiscovery_Technical.html" \
    "Service Discovery - Technical Documentation"

# Export User Manual
convert_md_to_html \
    "$SCRIPT_DIR/ServiceDiscovery_UserManual.md" \
    "$OUTPUT_DIR/ServiceDiscovery_UserManual.html" \
    "Service Discovery - User Manual"

# Create index page
cat > "$OUTPUT_DIR/index.html" <<'INDEX_HTML'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Service Discovery Documentation</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            max-width: 900px;
            margin: 0 auto;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
        }
        .container {
            background-color: white;
            padding: 40px;
            border-radius: 12px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.3);
        }
        h1 {
            color: #2c3e50;
            text-align: center;
            margin-bottom: 10px;
        }
        .subtitle {
            text-align: center;
            color: #7f8c8d;
            margin-bottom: 40px;
        }
        .docs-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
            margin-top: 30px;
        }
        .doc-card {
            border: 2px solid #ecf0f1;
            border-radius: 8px;
            padding: 30px;
            text-align: center;
            transition: all 0.3s ease;
            text-decoration: none;
            color: inherit;
            display: block;
        }
        .doc-card:hover {
            border-color: #3498db;
            transform: translateY(-5px);
            box-shadow: 0 5px 20px rgba(52, 152, 219, 0.3);
        }
        .doc-card h2 {
            color: #3498db;
            margin-top: 0;
        }
        .doc-card p {
            color: #7f8c8d;
            margin-bottom: 20px;
        }
        .doc-card .icon {
            font-size: 48px;
            margin-bottom: 20px;
        }
        .features {
            margin-top: 40px;
            background-color: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
        }
        .features h3 {
            color: #2c3e50;
            margin-top: 0;
        }
        .features ul {
            list-style-type: none;
            padding-left: 0;
        }
        .features li {
            padding: 8px 0;
            border-bottom: 1px solid #dee2e6;
        }
        .features li:before {
            content: "✓ ";
            color: #28a745;
            font-weight: bold;
        }
        .footer {
            text-align: center;
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #ddd;
            color: #888;
        }
        @media (max-width: 768px) {
            .docs-grid {
                grid-template-columns: 1fr;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Service Discovery & Failover</h1>
        <p class="subtitle">Complete Documentation for HelixTrack Core v1.0.0</p>

        <div class="docs-grid">
            <a href="ServiceDiscovery_Technical.html" class="doc-card">
                <div class="icon">📘</div>
                <h2>Technical Documentation</h2>
                <p>Architecture, API reference, database schema, and implementation details</p>
                <p><strong>For: Developers & Architects</strong></p>
            </a>

            <a href="ServiceDiscovery_UserManual.html" class="doc-card">
                <div class="icon">📗</div>
                <h2>User Manual</h2>
                <p>Step-by-step guides, best practices, and troubleshooting for operators</p>
                <p><strong>For: Operators & Administrators</strong></p>
            </a>
        </div>

        <div class="features">
            <h3>System Features</h3>
            <ul>
                <li>Dynamic Service Registration with cryptographic verification</li>
                <li>Automatic Health Monitoring (1-minute intervals)</li>
                <li>Automatic Failover to backup services</li>
                <li>Automatic Failback when primary recovers</li>
                <li>Secure Service Rotation with multi-layer verification</li>
                <li>Complete Audit Trail for all operations</li>
                <li>Priority-Based Service Selection</li>
                <li>Production-Ready with 100% test coverage</li>
            </ul>
        </div>

        <div class="footer">
            <p>Generated on GENERATION_DATE</p>
            <p>&copy; 2025 HelixTrack Core - Service Discovery System v1.0.0</p>
        </div>
    </div>
</body>
</html>
INDEX_HTML

sed -i "s/GENERATION_DATE/$(date '+%Y-%m-%d %H:%M:%S')/" "$OUTPUT_DIR/index.html"

echo ""
echo "========================================="
echo "Documentation export complete!"
echo "========================================="
echo ""
echo "Generated files:"
echo "  • $OUTPUT_DIR/index.html (start here)"
echo "  • $OUTPUT_DIR/ServiceDiscovery_Technical.html"
echo "  • $OUTPUT_DIR/ServiceDiscovery_UserManual.html"
echo ""
echo "To view, open: $OUTPUT_DIR/index.html"
echo ""
