/*
    LabelAssetMapping.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "LabelAssetMapping.h"

std::string LabelAssetMapping::getId() {    
    return this->id;
}

void LabelAssetMapping::setId(std::string value) {
    this->id = value;
}

std::string LabelAssetMapping::getLabelId() {    
    return this->labelId;
}

void LabelAssetMapping::setLabelId(std::string value) {
    this->labelId = value;
}

std::string LabelAssetMapping::getAssetId() {    
    return this->assetId;
}

void LabelAssetMapping::setAssetId(std::string value) {
    this->assetId = value;
}

int LabelAssetMapping::getCreated() {    
    return this->created;
}

void LabelAssetMapping::setCreated(int value) {
    this->created = value;
}

int LabelAssetMapping::getModified() {    
    return this->modified;
}

void LabelAssetMapping::setModified(int value) {
    this->modified = value;
}

bool LabelAssetMapping::isDeleted() {    
    return this->deleted;
}

void LabelAssetMapping::setDeleted(bool value) {
    this->deleted = value;
}

