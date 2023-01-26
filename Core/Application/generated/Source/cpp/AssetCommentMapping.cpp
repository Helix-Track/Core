/*
    AssetCommentMapping.cpp
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "AssetCommentMapping.h"

std::string AssetCommentMapping::getId() {    
    return this->id;
}

void AssetCommentMapping::setId(std::string &value) {
    this->id = value;
}

std::string AssetCommentMapping::getAssetId() {    
    return this->assetId;
}

void AssetCommentMapping::setAssetId(std::string &value) {
    this->assetId = value;
}

std::string AssetCommentMapping::getCommentId() {    
    return this->commentId;
}

void AssetCommentMapping::setCommentId(std::string &value) {
    this->commentId = value;
}

int AssetCommentMapping::getCreated() {    
    return this->created;
}

void AssetCommentMapping::setCreated(int &value) {
    this->created = value;
}

int AssetCommentMapping::getModified() {    
    return this->modified;
}

void AssetCommentMapping::setModified(int &value) {
    this->modified = value;
}

bool AssetCommentMapping::isDeleted() {    
    return this->deleted;
}

void AssetCommentMapping::setDeleted(bool &value) {
    this->deleted = value;
}

