/*
    LabelProjectMapping.cpp
    Generated with 'sql2code' 0.0.2
    https://github.com/red-elf/SQL-to-Code
*/

#include "LabelProjectMapping.h"

std::string LabelProjectMapping::getId() {    
    return this->id;
}

void LabelProjectMapping::setId(std::string &value) {
    this->id = value;
}

std::string LabelProjectMapping::getLabelId() {    
    return this->labelId;
}

void LabelProjectMapping::setLabelId(std::string &value) {
    this->labelId = value;
}

std::string LabelProjectMapping::getProjectId() {    
    return this->projectId;
}

void LabelProjectMapping::setProjectId(std::string &value) {
    this->projectId = value;
}

int LabelProjectMapping::getCreated() {    
    return this->created;
}

void LabelProjectMapping::setCreated(int &value) {
    this->created = value;
}

int LabelProjectMapping::getModified() {    
    return this->modified;
}

void LabelProjectMapping::setModified(int &value) {
    this->modified = value;
}

bool LabelProjectMapping::isDeleted() {    
    return this->deleted;
}

void LabelProjectMapping::setDeleted(bool &value) {
    this->deleted = value;
}

