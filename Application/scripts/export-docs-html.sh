#!/bin/bash
# Export documentation to HTML

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DOCS_DIR="$PROJECT_ROOT/docs"
HTML_DIR="$PROJECT_ROOT/docs/html"

echo "Exporting documentation to HTML..."

# Create HTML directory
mkdir -p "$HTML_DIR"

# Check if pandoc is installed
if ! command -v pandoc &> /dev/null; then
    echo "Warning: pandoc is not installed. Using simple Markdown to HTML conversion."
    echo "For better results, install pandoc: sudo apt-get install pandoc"
    echo ""

    # Simple conversion without pandoc
    for md_file in "$DOCS_DIR"/*.md; do
        if [ -f "$md_file" ]; then
            filename=$(basename "$md_file" .md)
            html_file="$HTML_DIR/${filename}.html"

            cat > "$html_file" << 'HTMLHEADER'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>HelixTrack Core Documentation</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6;
            max-width: 900px;
            margin: 0 auto;
            padding: 20px;
            color: #333;
        }
        h1, h2, h3, h4, h5, h6 {
            margin-top: 24px;
            margin-bottom: 16px;
            font-weight: 600;
            line-height: 1.25;
        }
        h1 { font-size: 2em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; }
        h2 { font-size: 1.5em; border-bottom: 1px solid #eaecef; padding-bottom: 0.3em; }
        h3 { font-size: 1.25em; }
        code {
            background-color: #f6f8fa;
            border-radius: 3px;
            font-size: 85%;
            margin: 0;
            padding: 0.2em 0.4em;
            font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
        }
        pre {
            background-color: #f6f8fa;
            border-radius: 3px;
            font-size: 85%;
            line-height: 1.45;
            overflow: auto;
            padding: 16px;
        }
        pre code {
            background-color: transparent;
            border: 0;
            display: inline;
            line-height: inherit;
            margin: 0;
            overflow: visible;
            padding: 0;
            word-wrap: normal;
        }
        table {
            border-collapse: collapse;
            width: 100%;
            margin: 16px 0;
        }
        table th, table td {
            border: 1px solid #dfe2e5;
            padding: 6px 13px;
        }
        table th {
            background-color: #f6f8fa;
            font-weight: 600;
        }
        blockquote {
            border-left: 4px solid #dfe2e5;
            color: #6a737d;
            padding: 0 1em;
            margin: 0;
        }
        a {
            color: #0366d6;
            text-decoration: none;
        }
        a:hover {
            text-decoration: underline;
        }
        .toc {
            background-color: #f6f8fa;
            border: 1px solid #dfe2e5;
            border-radius: 3px;
            padding: 16px;
            margin: 24px 0;
        }
    </style>
</head>
<body>
HTMLHEADER

            # Simple markdown to HTML (basic conversion)
            sed -e 's/^# \(.*\)/<h1>\1<\/h1>/' \
                -e 's/^## \(.*\)/<h2>\1<\/h2>/' \
                -e 's/^### \(.*\)/<h3>\1<\/h3>/' \
                -e 's/^#### \(.*\)/<h4>\1<\/h4>/' \
                -e 's/^```\(.*\)/<pre><code>/' \
                -e 's/^```$/<\/code><\/pre>/' \
                -e 's/`\([^`]*\)`/<code>\1<\/code>/g' \
                -e 's/^\* \(.*\)/<li>\1<\/li>/' \
                -e 's/^\- \(.*\)/<li>\1<\/li>/' \
                "$md_file" >> "$html_file"

            cat >> "$html_file" << 'HTMLFOOTER'
</body>
</html>
HTMLFOOTER

            echo "  Exported: $filename.html"
        fi
    done
else
    # Use pandoc for better conversion
    for md_file in "$DOCS_DIR"/*.md; do
        if [ -f "$md_file" ]; then
            filename=$(basename "$md_file" .md)
            html_file="$HTML_DIR/${filename}.html"

            pandoc "$md_file" \
                -f markdown \
                -t html5 \
                --standalone \
                --toc \
                --toc-depth=3 \
                --css=style.css \
                --metadata title="HelixTrack Core - $filename" \
                -o "$html_file"

            echo "  Exported: $filename.html"
        fi
    done

    # Create CSS file
    cat > "$HTML_DIR/style.css" << 'CSS'
body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    line-height: 1.6;
    max-width: 900px;
    margin: 0 auto;
    padding: 20px;
    color: #333;
}
h1, h2, h3, h4, h5, h6 {
    margin-top: 24px;
    margin-bottom: 16px;
    font-weight: 600;
    line-height: 1.25;
}
code {
    background-color: #f6f8fa;
    border-radius: 3px;
    font-size: 85%;
    padding: 0.2em 0.4em;
    font-family: "SFMono-Regular", Consolas, monospace;
}
pre {
    background-color: #f6f8fa;
    border-radius: 3px;
    padding: 16px;
    overflow: auto;
}
#TOC {
    background-color: #f6f8fa;
    border: 1px solid #dfe2e5;
    border-radius: 3px;
    padding: 16px;
    margin: 24px 0;
}
CSS
fi

# Create index page
cat > "$HTML_DIR/index.html" << 'INDEXHTML'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>HelixTrack Core Documentation</title>
    <link rel="stylesheet" href="style.css">
</head>
<body>
    <h1>HelixTrack Core Documentation</h1>
    <p>Welcome to the HelixTrack Core documentation.</p>

    <h2>Available Documentation</h2>
    <ul>
        <li><a href="USER_MANUAL.html">User Manual</a> - Complete guide for users</li>
        <li><a href="DEPLOYMENT.html">Deployment Guide</a> - Production deployment instructions</li>
    </ul>

    <h2>Quick Links</h2>
    <ul>
        <li><a href="../README.md">README</a></li>
        <li><a href="../../README.md">Project README</a></li>
    </ul>

    <footer style="margin-top: 50px; padding-top: 20px; border-top: 1px solid #eee; color: #666;">
        <p>HelixTrack Core v1.0.0 | Generated: $(date)</p>
    </footer>
</body>
</html>
INDEXHTML

echo ""
echo "Documentation exported to: $HTML_DIR"
echo "Open $HTML_DIR/index.html in your browser to view."
