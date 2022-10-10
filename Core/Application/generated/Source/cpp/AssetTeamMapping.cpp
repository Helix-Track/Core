/*
    AssetTeamMapping.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "AssetTeamMapping.h"

std::string AssetTeamMapping::getId() {    
    return this->id;
}

void AssetTeamMapping::setId(std::string &value) {
    this->id = value;
}

std::string AssetTeamMapping::getAssetId() {    
    return this->assetId;
}

void AssetTeamMapping::setAssetId(std::string &value) {
    this->assetId = value;
}

std::string AssetTeamMapping::getTeamId() {    
    return this->teamId;
}

void AssetTeamMapping::setTeamId(std::string &value) {
    this->teamId = value;
}

int AssetTeamMapping::getCreated() {    
    return this->created;
}

void AssetTeamMapping::setCreated(int &value) {
    this->created = value;
}

int AssetTeamMapping::getModified() {    
    return this->modified;
}

void AssetTeamMapping::setModified(int &value) {
    this->modified = value;
}

bool AssetTeamMapping::isDeleted() {    
    return this->deleted;
}

void AssetTeamMapping::setDeleted(bool &value) {
    this->deleted = value;
}

