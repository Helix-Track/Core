/*
    AssetProjectMapping.cpp
    Generated with 'sql2code' 0.0.2
    https://github.com/red-elf/SQL-to-Code
*/

#include "AssetProjectMapping.h"

std::string AssetProjectMapping::getId() {    
    return this->id;
}

void AssetProjectMapping::setId(std::string &value) {
    this->id = value;
}

std::string AssetProjectMapping::getAssetId() {    
    return this->assetId;
}

void AssetProjectMapping::setAssetId(std::string &value) {
    this->assetId = value;
}

std::string AssetProjectMapping::getProjectId() {    
    return this->projectId;
}

void AssetProjectMapping::setProjectId(std::string &value) {
    this->projectId = value;
}

int AssetProjectMapping::getCreated() {    
    return this->created;
}

void AssetProjectMapping::setCreated(int &value) {
    this->created = value;
}

int AssetProjectMapping::getModified() {    
    return this->modified;
}

void AssetProjectMapping::setModified(int &value) {
    this->modified = value;
}

bool AssetProjectMapping::isDeleted() {    
    return this->deleted;
}

void AssetProjectMapping::setDeleted(bool &value) {
    this->deleted = value;
}

