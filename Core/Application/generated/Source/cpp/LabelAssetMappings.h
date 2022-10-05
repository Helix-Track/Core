/*
    LabelAssetMappings.h
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class LabelAssetMappings {

private:
    std::string id;
    std::string labelId;
    std::string assetId;
    int created;
    int modified;
    bool deleted;

public:
    std::string getId();
    void setId(std::string value);
    std::string getLabelId();
    void setLabelId(std::string value);
    std::string getAssetId();
    void setAssetId(std::string value);
    int getCreated();
    void setCreated(int value);
    int getModified();
    void setModified(int value);
    bool isDeleted();
    void setDeleted(bool value);
};