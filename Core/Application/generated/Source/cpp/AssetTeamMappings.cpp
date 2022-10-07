/*
    AssetTeamMappings.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "AssetTeamMappings.h"

std::string AssetTeamMappings::getId() {    
    return this->id;
}

void AssetTeamMappings::setId(std::string value) {
    this->id = value;
}

std::string AssetTeamMappings::getAssetId() {    
    return this->assetId;
}

void AssetTeamMappings::setAssetId(std::string value) {
    this->assetId = value;
}

std::string AssetTeamMappings::getTeamId() {    
    return this->teamId;
}

void AssetTeamMappings::setTeamId(std::string value) {
    this->teamId = value;
}

int AssetTeamMappings::getCreated() {    
    return this->created;
}

void AssetTeamMappings::setCreated(int value) {
    this->created = value;
}

int AssetTeamMappings::getModified() {    
    return this->modified;
}

void AssetTeamMappings::setModified(int value) {
    this->modified = value;
}

bool AssetTeamMappings::isDeleted() {    
    return this->deleted;
}

void AssetTeamMappings::setDeleted(bool value) {
    this->deleted = value;
}

