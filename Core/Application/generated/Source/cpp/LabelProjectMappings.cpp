/*
    LabelProjectMappings.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "LabelProjectMappings.h"

std::string LabelProjectMappings::getId() {    
    return this->id;
}

void LabelProjectMappings::setId(std::string value) {
    this->id = value;
}

std::string LabelProjectMappings::getLabelId() {    
    return this->labelId;
}

void LabelProjectMappings::setLabelId(std::string value) {
    this->labelId = value;
}

std::string LabelProjectMappings::getProjectId() {    
    return this->projectId;
}

void LabelProjectMappings::setProjectId(std::string value) {
    this->projectId = value;
}

int LabelProjectMappings::getCreated() {    
    return this->created;
}

void LabelProjectMappings::setCreated(int value) {
    this->created = value;
}

int LabelProjectMappings::getModified() {    
    return this->modified;
}

void LabelProjectMappings::setModified(int value) {
    this->modified = value;
}

bool LabelProjectMappings::isDeleted() {    
    return this->deleted;
}

void LabelProjectMappings::setDeleted(bool value) {
    this->deleted = value;
}

