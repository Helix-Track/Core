/*
    WorkflowSteps.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "WorkflowSteps.h"

std::string WorkflowSteps::getId() {    
    return this->id;
}

void WorkflowSteps::setId(std::string value) {
    this->id = value;
}

std::string WorkflowSteps::getTitle() {    
    return this->title;
}

void WorkflowSteps::setTitle(std::string value) {
    this->title = value;
}

std::string WorkflowSteps::getDescription() {    
    return this->description;
}

void WorkflowSteps::setDescription(std::string value) {
    this->description = value;
}

std::string WorkflowSteps::getWorkflowId() {    
    return this->workflowId;
}

void WorkflowSteps::setWorkflowId(std::string value) {
    this->workflowId = value;
}

std::string WorkflowSteps::getWorkflowStepId() {    
    return this->workflowStepId;
}

void WorkflowSteps::setWorkflowStepId(std::string value) {
    this->workflowStepId = value;
}

std::string WorkflowSteps::getTicketStatusId() {    
    return this->ticketStatusId;
}

void WorkflowSteps::setTicketStatusId(std::string value) {
    this->ticketStatusId = value;
}

int WorkflowSteps::getCreated() {    
    return this->created;
}

void WorkflowSteps::setCreated(int value) {
    this->created = value;
}

int WorkflowSteps::getModified() {    
    return this->modified;
}

void WorkflowSteps::setModified(int value) {
    this->modified = value;
}

bool WorkflowSteps::isDeleted() {    
    return this->deleted;
}

void WorkflowSteps::setDeleted(bool value) {
    this->deleted = value;
}

