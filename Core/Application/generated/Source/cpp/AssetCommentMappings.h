/*
    AssetCommentMappings.h
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class AssetCommentMappings {

private:
    std::string id;
    std::string assetId;
    std::string commentId;
    int created;
    int modified;
    bool deleted;

public:
    std::string getId();
    void setId(std::string value);
    std::string getAssetId();
    void setAssetId(std::string value);
    std::string getCommentId();
    void setCommentId(std::string value);
    int getCreated();
    void setCreated(int value);
    int getModified();
    void setModified(int value);
    bool isDeleted();
    void setDeleted(bool value);
};