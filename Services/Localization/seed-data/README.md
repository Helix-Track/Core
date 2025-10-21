# Localization Seed Data

This directory contains the seed data for initializing the HelixTrack Localization service with default languages and translations.

## Directory Structure

```
seed-data/
├── languages.json              # Language definitions (10 languages)
├── localization-keys.json      # Localization keys with metadata (79 keys)
├── localizations/              # Translation files per language
│   ├── en.json                 # English (100% complete - 79 strings)
│   ├── de.json                 # German (100% complete - 79 strings)
│   ├── fr.json                 # French (100% complete - 79 strings)
│   ├── es.json                 # Spanish (placeholder)
│   ├── pt.json                 # Portuguese (placeholder)
│   ├── ru.json                 # Russian (placeholder)
│   ├── zh.json                 # Chinese (placeholder)
│   ├── ja.json                 # Japanese (placeholder)
│   ├── ar.json                 # Arabic (placeholder - RTL)
│   └── he.json                 # Hebrew (placeholder - RTL)
└── README.md                   # This file
```

## Data Files

### languages.json

Defines 10 languages with the following structure:

```json
{
  "code": "en",              // ISO 639-1 language code
  "name": "English",         // English name
  "native_name": "English",  // Native name
  "is_rtl": false,           // Right-to-left flag
  "is_active": true,         // Enabled/disabled flag
  "is_default": true         // Default language flag (only one)
}
```

**Supported Languages:**
- English (en) - Default, 100% complete
- German (de) - 100% complete
- French (fr) - 100% complete
- Spanish (es) - Placeholder
- Portuguese (pt) - Placeholder
- Russian (ru) - Placeholder
- Chinese (zh) - Placeholder
- Japanese (ja) - Placeholder
- Arabic (ar) - RTL, Placeholder
- Hebrew (he) - RTL, Placeholder

### localization-keys.json

Defines all localization keys with metadata:

```json
{
  "key": "error.invalid_request",           // Unique key
  "category": "error",                      // Category for organization
  "description": "Invalid request error",   // Developer notes
  "context": "api_validation",              // Usage context
  "variables": []                           // Variable placeholders
}
```

**Categories:**
- `error` - Error messages (32 keys)
- `common` - Common UI elements (11 keys)
- `navigation` - Navigation items (8 keys)
- `authentication` - Auth-related strings (6 keys)
- `dashboard` - Dashboard strings (4 keys)
- `project` - Project-related strings (4 keys)
- `ticket` - Ticket-related strings (7 keys)
- `settings` - Settings strings (5 keys)
- `application` - Application-level strings (2 keys)

**Total:** 79 localization keys

### localizations/*.json

Translation files for each language, mapping keys to translated strings:

```json
{
  "error.invalid_request": "Invalid request",
  "common.ok": "OK",
  "app.welcome_user": "Welcome to HelixTrack, {name}!"
}
```

**Variable Interpolation:**
Keys with variables (e.g., `{name}`, `{count}`) support runtime value replacement.

## Usage

### Automatic Population (Startup)

The Localization service automatically loads this seed data on first startup if the database is empty:

```bash
cd Core/Services/Localization
./htLocalization --config=configs/default.json
```

On startup, the service will:
1. Check if database is populated
2. If empty, load seed data from this directory
3. Import languages, keys, and localizations
4. Build and cache catalogs
5. Log summary to console

### Manual Import via Script

```bash
cd Core/Services/Localization
./scripts/populate-from-seed.sh
```

### Manual Import via API

```bash
curl -X POST https://localhost:8085/v1/admin/import \
  -H "Authorization: Bearer YOUR_ADMIN_JWT" \
  -H "Content-Type: application/json" \
  -d @seed-data/import-payload.json
```

## Adding New Translations

### Step 1: Add Language

Edit `languages.json` and add new language:

```json
{
  "code": "it",
  "name": "Italian",
  "native_name": "Italiano",
  "is_rtl": false,
  "is_active": true,
  "is_default": false
}
```

### Step 2: Create Translation File

Create `localizations/it.json` with translations:

```json
{
  "error.success": "Successo",
  "error.invalid_request": "Richiesta non valida",
  ...
}
```

### Step 3: Import

Re-run the population script or use the import API endpoint.

## Adding New Keys

### Step 1: Add Key Definition

Edit `localization-keys.json`:

```json
{
  "key": "feature.new_feature",
  "category": "feature",
  "description": "New feature message",
  "context": "feature_page",
  "variables": []
}
```

### Step 2: Add Translations

Add the key to each language file in `localizations/`:

**en.json:**
```json
{
  "feature.new_feature": "Welcome to our new feature!"
}
```

**de.json:**
```json
{
  "feature.new_feature": "Willkommen zu unserer neuen Funktion!"
}
```

### Step 3: Import

Re-run the population script or use the import API endpoint.

## Backup & Export

### Export Current Database to Seed Format

```bash
cd Core/Services/Localization
./scripts/export-to-seed.sh
```

This will:
1. Export all languages to `languages.json`
2. Export all keys to `localization-keys.json`
3. Export all translations to `localizations/*.json`
4. Create backup with timestamp

### Periodic Backup

The service automatically backs up to `backups/` directory:
- Hourly: Incremental backups
- Daily: Full backups
- Weekly: Full backups with version tags

## Source of Translations

These seed translations were collected from:
- **Core Application** (`Core/Application/internal/models/errors.go`) - 25 error messages
- **Web Client** (`Web-Client/src/assets/i18n/en.json`) - 35 error messages
- **Desktop Client** (`Desktop-Client/src/assets/i18n/en.json`) - 35 error messages
- **Android Client** (`Android-Client/app/src/main/res/values/strings.xml`) - 103 UI strings
- **Core Handlers** (`Core/Application/internal/handlers/*.go`) - 200+ handler messages

All hardcoded strings have been extracted and centralized into this localization system.

## Translation Coverage

| Language | Code | Completion | Notes |
|----------|------|------------|-------|
| English | en | 100% (79/79) | Default language |
| German | de | 100% (79/79) | Complete |
| French | fr | 100% (79/79) | Complete |
| Spanish | es | 0% (0/79) | Placeholder - needs translation |
| Portuguese | pt | 0% (0/79) | Placeholder - needs translation |
| Russian | ru | 0% (0/79) | Placeholder - needs translation |
| Chinese | zh | 0% (0/79) | Placeholder - needs translation |
| Japanese | ja | 0% (0/79) | Placeholder - needs translation |
| Arabic | ar | 0% (0/79) | Placeholder - needs translation (RTL) |
| Hebrew | he | 0% (0/79) | Placeholder - needs translation (RTL) |

## Contributing Translations

To contribute translations for placeholder languages:

1. Copy `localizations/en.json` to `localizations/XX.json` (XX = language code)
2. Translate all values (keep keys unchanged)
3. Test variable interpolation (e.g., `{name}`, `{count}`)
4. Submit pull request or use admin UI

For RTL languages (Arabic, Hebrew):
- UI components will automatically apply RTL layout
- Text direction is handled by the client applications
- No special formatting needed in translation files

## Version History

- **v1.0.0** (2025-01-15): Initial seed data
  - 10 languages defined
  - 79 localization keys
  - 3 complete translations (en, de, fr)
  - 7 placeholder languages

---

**Maintained by:** HelixTrack Localization Team
**Last Updated:** 2025-01-15
