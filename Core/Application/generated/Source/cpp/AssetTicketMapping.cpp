/*
    AssetTicketMapping.cpp
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "AssetTicketMapping.h"

std::string AssetTicketMapping::getId() {    
    return this->id;
}

void AssetTicketMapping::setId(std::string &value) {
    this->id = value;
}

std::string AssetTicketMapping::getAssetId() {    
    return this->assetId;
}

void AssetTicketMapping::setAssetId(std::string &value) {
    this->assetId = value;
}

std::string AssetTicketMapping::getTicketId() {    
    return this->ticketId;
}

void AssetTicketMapping::setTicketId(std::string &value) {
    this->ticketId = value;
}

int AssetTicketMapping::getCreated() {    
    return this->created;
}

void AssetTicketMapping::setCreated(int &value) {
    this->created = value;
}

int AssetTicketMapping::getModified() {    
    return this->modified;
}

void AssetTicketMapping::setModified(int &value) {
    this->modified = value;
}

bool AssetTicketMapping::isDeleted() {    
    return this->deleted;
}

void AssetTicketMapping::setDeleted(bool &value) {
    this->deleted = value;
}

