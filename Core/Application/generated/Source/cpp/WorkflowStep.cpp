/*
    WorkflowStep.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "WorkflowStep.h"

std::string WorkflowStep::getId() {    
    return this->id;
}

void WorkflowStep::setId(std::string &value) {
    this->id = value;
}

std::string WorkflowStep::getTitle() {    
    return this->title;
}

void WorkflowStep::setTitle(std::string &value) {
    this->title = value;
}

std::string WorkflowStep::getDescription() {    
    return this->description;
}

void WorkflowStep::setDescription(std::string &value) {
    this->description = value;
}

std::string WorkflowStep::getWorkflowId() {    
    return this->workflowId;
}

void WorkflowStep::setWorkflowId(std::string &value) {
    this->workflowId = value;
}

std::string WorkflowStep::getWorkflowStepId() {    
    return this->workflowStepId;
}

void WorkflowStep::setWorkflowStepId(std::string &value) {
    this->workflowStepId = value;
}

std::string WorkflowStep::getTicketStatusId() {    
    return this->ticketStatusId;
}

void WorkflowStep::setTicketStatusId(std::string &value) {
    this->ticketStatusId = value;
}

int WorkflowStep::getCreated() {    
    return this->created;
}

void WorkflowStep::setCreated(int &value) {
    this->created = value;
}

int WorkflowStep::getModified() {    
    return this->modified;
}

void WorkflowStep::setModified(int &value) {
    this->modified = value;
}

bool WorkflowStep::isDeleted() {    
    return this->deleted;
}

void WorkflowStep::setDeleted(bool &value) {
    this->deleted = value;
}

